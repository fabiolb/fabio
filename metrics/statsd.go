package metrics

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

type StatsdRegistry struct {
	addr   string
	prefix string
	conn   net.Conn
}

func NewStatsdRegistry(prefix, addr string) *StatsdRegistry {
	client := StatsdRegistry{prefix: prefix, addr: addr}
	client.Open()
	return &client
}

func (client *StatsdRegistry) Open() {
	conn, err := net.Dial("udp", client.addr)
	if err != nil {
		log.Println(err)
	}
	client.conn = conn
}

func (client *StatsdRegistry) Close() {
	client.conn.Close()
}

// Names is not supported by Statsd.
func (client *StatsdRegistry) Names() []string { return nil }

// Unregister is implicitly supported by Statsd,
// stop submitting the metric and it stops being sent to Statsd.
func (client *StatsdRegistry) Unregister(name string) {}

// UnregisterAll is implicitly supported by Statsd,
// stop submitting metrics and they will no longer be sent to Statsd.
func (client *StatsdRegistry) UnregisterAll() {}

// Arbitrarily updates a list of stats by a delta
func (client *StatsdRegistry) UpdateStats(stats []string, delta int, sampleRate float32) {
	statsToSend := make(map[string]string)
	for _, stat := range stats {
		updateString := fmt.Sprintf("%d|c", delta)
		statsToSend[stat] = updateString
	}
	client.Send(statsToSend, sampleRate)
}

// Sends data to udp statsd daemon
func (client *StatsdRegistry) Send(data map[string]string, sampleRate float32) {
	sampledData := make(map[string]string)
	if sampleRate < 1 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		rNum := r.Float32()
		if rNum <= sampleRate {
			for stat, value := range data {
				sampledUpdateString := fmt.Sprintf("%s|@%f", value, sampleRate)
				sampledData[stat] = sampledUpdateString
			}
		}
	} else {
		sampledData = data
	}

	for k, v := range sampledData {
		updateString := fmt.Sprintf("%s:%s", k, v)
		_, err := fmt.Fprintf(client.conn, updateString)
		if err != nil {
			log.Println(err)
		}
	}
}

func (client StatsdRegistry) GetCounter(name string) Counter {
	return StatsdCounter{&client, client.prefix + name}
}

type StatsdCounter struct {
	client *StatsdRegistry
	name   string
}

func (c StatsdCounter) Inc(n int64) {
	stats := []string{c.name}
	c.client.UpdateStats(stats, int(n), 1)
}

func (client StatsdRegistry) GetTimer(name string) Timer {
	return StatsdTimer{&client, client.prefix + name}
}

type StatsdTimer struct {
	client *StatsdRegistry
	name   string
}

func (t StatsdTimer) Update(n time.Duration) {
	duration := int64(n / time.Millisecond)
	updateString := fmt.Sprintf("%d|ms", duration)
	stats := map[string]string{t.name: updateString}
	t.client.Send(stats, 1)
}

func (t StatsdTimer) UpdateSince(then time.Time) {
	duration := time.Since(then)
	t.Update(duration)
}

// Percentile is not supported by Statsd.
func (t StatsdTimer) Percentile(nth float64) float64 { return 0 }

// Rate1 is not supported by Statsd.
func (t StatsdTimer) Rate1() float64 { return 0 }
