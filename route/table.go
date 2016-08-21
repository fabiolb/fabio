package route

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/eBay/fabio/metrics"
)

var errInvalidPrefix = errors.New("route: prefix must not be empty")
var errInvalidTarget = errors.New("route: target must not be empty")
var errNoMatch = errors.New("route: no target match")

// table stores the active routing table. Must never be nil.
var table atomic.Value

// ServiceRegistry stores the metrics for the services.
var ServiceRegistry metrics.Registry = metrics.NoopRegistry{}

// init initializes the routing table.
func init() {
	table.Store(make(Table))
}

// GetTable returns the active routing table. The function
// is safe to be called from multiple goroutines and the
// value is never nil.
func GetTable() Table {
	return table.Load().(Table)
}

// mu guards table and registry in SetTable.
var mu sync.Mutex

// SetTable sets the active routing table. A nil value
// logs a warning and is ignored. The function is safe
// to be called from multiple goroutines.
func SetTable(t Table) {
	if t == nil {
		log.Print("[WARN] Ignoring nil routing table")
		return
	}
	mu.Lock()
	table.Store(t)
	syncRegistry(t)
	mu.Unlock()
	log.Printf("[INFO] Updated config to\n%s", t)
}

// syncRegistry unregisters all inactive timers.
// It assumes that all timers of the table have
// already been registered.
func syncRegistry(t Table) {
	timers := map[string]bool{}

	// get all registered timers
	for _, name := range ServiceRegistry.Names() {
		timers[name] = false
	}

	// mark the ones from this table as active.
	// this can also add new entries but we do not
	// really care since we are only interested in the
	// inactive ones.
	for _, routes := range t {
		for _, r := range routes {
			for _, tg := range r.Targets {
				timers[tg.timerName] = true
			}
		}
	}

	// unregister inactive timers
	for name, active := range timers {
		if !active {
			ServiceRegistry.Unregister(name)
			log.Printf("[INFO] Unregistered timer %s", name)
		}
	}
}

// Table contains a set of routes grouped by host.
// The host routes are sorted from most to least specific
// by sorting the routes in reverse order by path.
type Table map[string]Routes

// hostpath splits a host/path prefix into a host and a path.
// The path always starts with a slash
func hostpath(prefix string) (host string, path string) {
	p := strings.SplitN(prefix, "/", 2)
	host, path = p[0], ""
	if len(p) == 1 {
		return p[0], "/"
	}
	return p[0], "/" + p[1]
}

// AddRoute adds a new route prefix -> target for the given service.
func (t Table) AddRoute(service, prefix, target string, weight float64, tags []string) error {
	host, path := hostpath(prefix)

	if prefix == "" {
		return errInvalidPrefix
	}

	if target == "" {
		return errInvalidTarget
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("route: invalid target. %s", err)
	}

	r := newRoute(host, path)
	r.addTarget(service, targetURL, weight, tags)

	// add new host
	if t[host] == nil {
		t[host] = Routes{r}
		return nil
	}

	// add new route to existing host
	if t[host].find(path) == nil {
		t[host] = append(t[host], r)
		sort.Sort(t[host])
		return nil
	}

	// add new target to existing route
	t[host].find(path).addTarget(service, targetURL, weight, tags)

	return nil
}

func (t Table) AddRouteWeight(service, prefix string, weight float64, tags []string) error {
	host, path := hostpath(prefix)

	if prefix == "" {
		return errInvalidPrefix
	}

	if t[host] == nil || t[host].find(path) == nil {
		return errNoMatch
	}

	if n := t[host].find(path).setWeight(service, weight, tags); n == 0 {
		return errNoMatch
	}
	return nil
}

// DelRoute removes one or more routes depending on the arguments.
// If service, prefix and target are provided then only this route
// is removed. Are only service and prefix provided then all routes
// for this service and prefix are removed. This removes all active
// instances of the service from the route. If only the service is
// provided then all routes for this service are removed. The service
// will no longer receive traffic. Routes with no targets are removed.
func (t Table) DelRoute(service, prefix, target string) error {
	switch {
	case prefix == "" && target == "":
		for _, routes := range t {
			for _, r := range routes {
				r.delService(service)
			}
		}

	case target == "":
		r := t.route(hostpath(prefix))
		if r == nil {
			return nil
		}
		r.delService(service)

	default:
		targetURL, err := url.Parse(target)
		if err != nil {
			return fmt.Errorf("route: invalid target. %s", err)
		}

		r := t.route(hostpath(prefix))
		if r == nil {
			return nil
		}
		r.delTarget(service, targetURL)
	}

	// remove all routes without targets
	for host, routes := range t {
		var clone Routes
		for _, r := range routes {
			if len(r.Targets) == 0 {
				continue
			}
			clone = append(clone, r)
		}
		t[host] = clone
	}

	// remove all hosts without routes
	for host, routes := range t {
		if len(routes) == 0 {
			delete(t, host)
		}
	}

	return nil
}

// route finds the route for host/path or returns nil if none exists.
func (t Table) route(host, path string) *Route {
	routes := t[host]
	if routes == nil {
		return nil
	}
	return routes.find(path)
}

// normalizeHost returns the hostname from the request
// and removes the default port if present.
func normalizeHost(req *http.Request) string {
	host := strings.ToLower(req.Host)
	if req.TLS == nil && strings.HasSuffix(host, ":80") {
		return host[:len(host)-3]
	}
	if req.TLS != nil && strings.HasSuffix(host, ":443") {
		return host[:len(host)-4]
	}
	return host
}

// Lookup finds a target url based on the current matcher and picker
// or nil if there is none. It first checks the routes for the host
// and if none matches then it falls back to generic routes without
// a host. This is useful for a catch-all '/' rule.
func (t Table) Lookup(req *http.Request, trace string) *Target {
	if trace != "" {
		if len(trace) > 16 {
			trace = trace[:15]
		}
		log.Printf("[TRACE] %s Tracing %s%s", trace, req.Host, req.RequestURI)
	}

	target := t.lookup(normalizeHost(req), req.RequestURI, trace)
	if target == nil {
		target = t.lookup("", req.RequestURI, trace)
	}

	if target != nil && trace != "" {
		log.Printf("[TRACE] %s Routing to service %s on %s", trace, target.Service, target.URL)
	}

	return target
}

func (t Table) LookupHost(host string) *Target {
	return t.lookup(host, "/", "")
}

func (t Table) lookup(host, path, trace string) *Target {
	for _, r := range t[host] {
		if match(path, r) {
			n := len(r.Targets)
			if n == 0 {
				return nil
			}

			var target *Target
			if n == 1 {
				target = r.Targets[0]
			} else {
				target = pick(r)
			}
			if trace != "" {
				log.Printf("[TRACE] %s Match %s%s", trace, r.Host, r.Path)
			}
			return target
		}
		if trace != "" {
			log.Printf("[TRACE] %s No match %s%s", trace, r.Host, r.Path)
		}
	}
	return nil
}

func (t Table) Config(addWeight bool) []string {
	var hosts []string
	for host := range t {
		if host != "" {
			hosts = append(hosts, host)
		}
	}
	sort.Strings(hosts)

	// entries without host come last
	hosts = append(hosts, "")

	var cfg []string
	for _, host := range hosts {
		for _, routes := range t[host] {
			cfg = append(cfg, routes.config(addWeight)...)
		}
	}
	return cfg
}

// String returns the routing table as config file which can
// be read by Parse() again.
func (t Table) String() string {
	return strings.Join(t.Config(false), "\n")
}
