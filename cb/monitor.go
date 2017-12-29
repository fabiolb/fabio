package cb

import (
	"log"
	"strings"
	"time"

	circuit "github.com/rubyist/circuitbreaker"
)

// Monitor implements a circuit breaker monitor which manages
// multiple circuit breakers and generates routing table updates
// from the state changes.
//
// Monitor generates a new routing table fragment on a state change
// and on a regular basis. The default is to trigger the breaker
// after three consecutive failures.
type Monitor struct {
	UpdateInterval time.Duration
	ConsecFailures int
	routes         chan string
	fail           chan string
	success        chan string
	done           chan struct{}
}

func NewMonitor() *Monitor {
	return &Monitor{
		UpdateInterval: 15 * time.Second,
		ConsecFailures: 3,
		routes:         make(chan string, 1),
		fail:           make(chan string, 100),
		success:        make(chan string, 100),
		done:           make(chan struct{}),
	}
}

func (m *Monitor) Stop() {
	close(m.done)
}

func (m *Monitor) Start() {
	cbs := make(map[string]*circuit.Breaker)

	getcb := func(addr string) *circuit.Breaker {
		cb := cbs[addr]
		if cb == nil {
			cb = circuit.NewConsecutiveBreaker(int64(m.ConsecFailures))
			cbs[addr] = cb
		}
		return cb
	}

	ticker := time.NewTicker(m.UpdateInterval)
	for {
		select {
		case <-ticker.C:
			ready := 0
			for addr, cb := range cbs {
				if cb.Tripped() && cb.Ready() {
					ready++
					log.Printf("[INFO] breaker: retrying routes for %s", addr)
				}
			}
			if ready > 0 {
				m.routes <- m.update(cbs)
			}

		case addr := <-m.fail:
			cb := getcb(addr)
			wasready := cb.Ready()
			cb.Fail()
			if wasready && cb.Tripped() {
				log.Printf("[WARN] breaker: breaker for %s tripped", addr)
				m.routes <- m.update(cbs)
			}

		case addr := <-m.success:
			cb := getcb(addr)
			wasready := cb.Ready()
			cb.Success()
			if !wasready && !cb.Tripped() {
				log.Printf("[INFO] breaker: breaker for %s recovered", addr)
				m.routes <- m.update(cbs)
			}

		case <-m.done:
			return
		}
	}
}

func (m *Monitor) SuccessHost(addr string) {
	select {
	case m.success <- addr:
	default:
	}
}

func (m *Monitor) FailHost(addr string) {
	select {
	case m.fail <- addr:
	default:
	}
}

func (m *Monitor) Routes() <-chan string {
	return m.routes
}

func (m *Monitor) update(cbs map[string]*circuit.Breaker) string {
	var s []string
	for addr, cb := range cbs {
		if cb.Tripped() || !cb.Ready() {
			s = append(s, "route del * * http://"+addr)
		}
	}
	routes := strings.Join(s, "\n")
	return routes
}
