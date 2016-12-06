package route

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"path"
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

// use new parser
var Parse = ParseNew

func ParseTable(s string) (Table, error) {
	defs, err := Parse(s)
	if err != nil {
		return nil, err
	}
	return BuildTable(defs)
}

func BuildTable(defs []*RouteDef) (t Table, err error) {
	t = Table{}
	for _, d := range defs {
		switch d.Cmd {
		case RouteAddCmd:
			err = t.AddRoute(d)
		case RouteDelCmd:
			err = t.DelRoute(d)
		case RouteWeightCmd:
			err = t.RouteWeight(d)
		default:
			err = fmt.Errorf("route: invalid command: %s", d.Cmd)
		}
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

// AddRoute adds a new route prefix -> target for the given service.
func (t Table) AddRoute(d *RouteDef) error {
	host, path := hostpath(d.Src)

	if d.Src == "" {
		return errInvalidPrefix
	}

	if d.Dst == "" {
		return errInvalidTarget
	}

	targetURL, err := url.Parse(d.Dst)
	if err != nil {
		return fmt.Errorf("route: invalid target. %s", err)
	}

	switch {
	// add new host
	case t[host] == nil:
		r := &Route{Host: host, Path: path, Opts: d.Opts}
		r.addTarget(d.Service, targetURL, d.Weight, d.Tags)
		t[host] = Routes{r}

	// add new route to existing host
	case t[host].find(path) == nil:
		r := &Route{Host: host, Path: path, Opts: d.Opts}
		r.addTarget(d.Service, targetURL, d.Weight, d.Tags)
		t[host] = append(t[host], r)
		sort.Sort(t[host])

	// add new target to existing route
	default:
		t[host].find(path).addTarget(d.Service, targetURL, d.Weight, d.Tags)
	}

	return nil
}

func (t Table) RouteWeight(d *RouteDef) error {
	host, path := hostpath(d.Src)

	if d.Src == "" {
		return errInvalidPrefix
	}

	if t[host] == nil || t[host].find(path) == nil {
		return errNoMatch
	}

	if n := t[host].find(path).setWeight(d.Service, d.Weight, d.Tags); n == 0 {
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
func (t Table) DelRoute(d *RouteDef) error {
	switch {
	case d.Src == "" && d.Dst == "":
		for _, routes := range t {
			for _, r := range routes {
				r.delService(d.Service)
			}
		}

	case d.Dst == "":
		r := t.route(hostpath(d.Src))
		if r == nil {
			return nil
		}
		r.delService(d.Service)

	default:
		targetURL, err := url.Parse(d.Dst)
		if err != nil {
			return fmt.Errorf("route: invalid target. %s", err)
		}

		r := t.route(hostpath(d.Src))
		if r == nil {
			return nil
		}
		r.delTarget(d.Service, targetURL)
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

// performs a match of path against a specific host in the Table map
// returns a matching route only
func (t Table) matchHostPathRoute(host string, path string, trace string) *Route {
	for _, r := range t[host] {
		if match(path, r) {
			return r
		}
		if trace != "" {
			log.Printf("[TRACE] %s No match %s%s", trace, r.Host, r.Path)
		}
	}

	return nil
}


// performs a match of path against a specific host in the Table map
func (t Table) matchHostPath(host string, path string, trace string) *Target {
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


// 1. if no host defined, just does a path match
// 2. if host defined does a glob host match, then paths against each match
func (t Table) lookup(host, lookupPath, trace string) *Target {
	// If no host defined, then it is just path matching
	if host == "" {
		return t.matchHostPath(host, lookupPath, trace)
	}

	routes := Routes{}

	// If there is a direct match, add that route as the first match
	var foundRoute = t.matchHostPathRoute(host, lookupPath, trace)
	if foundRoute != nil {
		routes = append(routes, foundRoute)
	}

	// TODO give an option for regex or glob host configuration
	// think about how to do regex whether to inject ^$ etc.

	// If the user wants to run regex search (TODO, do regex and make configuration)
	// Create a new array of targets

	// If not found against the host, perform a glob search against the hosts
	// then for each of those, run a path check, and add that entry into the routing table
	for hostKey, _ := range t {
		// If this matches a glob match, add it to the list of hosts to search
		var hasMatch, err = path.Match(hostKey, host)

		if err != nil {
			log.Printf("[ERROR] Glob matching error %s for host %s vs %s", err, hostKey, host)
		}

		// If the glob matches, then perform a search for the paths against the host found
		if hasMatch {
			var matchingRoute = t.matchHostPathRoute(hostKey, lookupPath, trace)
			if matchingRoute != nil {
				// Should only check for unique route (from above foundRoute)?
				routes = append(routes, matchingRoute)
			}
		}
	}

	// at this point iterate through the matching route to find the one that has the strongest match
	// this should be a method elsewhere, because it is identical to matchHostPath()
	sort.Sort(routes)
        for _, r := range routes {
		if match(lookupPath, r) {
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
