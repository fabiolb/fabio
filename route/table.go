package route

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync/atomic"
)

// active stores the current routing table. Should never be nil.
var active atomic.Value

var errInvalidPrefix = errors.New("route: prefix must not be empty")
var errInvalidTarget = errors.New("route: target must not be empty")
var errNoMatch = errors.New("route: no target match")

func init() {
	active.Store(make(Table))
}

func GetTable() Table {
	return active.Load().(Table)
}

func SetTable(t Table) {
	if t == nil {
		log.Print("[WARN] Ignoring nil routing table")
		return
	}
	active.Store(t)
	log.Printf("[INFO] Updated config to\n%s", t)
}

// Table contains a set of routes grouped by host.
// The host routes are sorted from most to least specific
// by sorting the routes in reverse order by path.
type Table map[string]routes

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
		t[host] = routes{r}
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

	if n := t[host].find(path).setWeight(weight, tags); n == 0 {
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
// will no longer receive traffic.
func (t Table) DelRoute(service, prefix, target string) error {
	switch {
	case prefix == "" && target == "":
		for _, hr := range t {
			for _, r := range hr {
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

	return nil
}

// route finds the route for host/path or returns nil if none exists.
func (t Table) route(host, path string) *route {
	hr := t[host]
	if hr == nil {
		return nil
	}
	return hr.find(path)
}

// lookup finds a target url based on the current matcher and picker
// or nil if there is none. It first checks the routes for the host
// and if none matches then it falls back to generic routes without
// a host. This is useful for a catch-all '/' rule.
func (t Table) lookup(req *http.Request, trace string) *target {
	if trace != "" {
		if len(trace) > 16 {
			trace = trace[:15]
		}
		log.Printf("[TRACE] %s Tracing %s%s", trace, req.Host, req.RequestURI)
	}

	u := t.doLookup(strings.ToLower(req.Host), req.RequestURI, trace)
	if u == nil {
		u = t.doLookup("", req.RequestURI, trace)
	}

	if trace != "" {
		log.Printf("[TRACE] %s Routing to %s", trace, u.URL)
	}

	return u
}

func (t Table) doLookup(host, path, trace string) *target {
	hr := t[host]
	if hr == nil {
		return nil
	}

	for _, r := range hr {
		if match(path, r) {
			n := len(r.targets)
			if n == 0 {
				return nil
			}
			if n == 1 {
				return r.targets[0]
			}
			if trace != "" {
				log.Printf("[TRACE] %s Match %s%s", trace, r.host, r.path)
			}
			return pick(r)
		}
		if trace != "" {
			log.Printf("[TRACE] %s No match %s%s", trace, r.host, r.path)
		}
	}
	return nil
}

func (t Table) Config(addWeight bool) []string {
	var hosts []string
	for h := range t {
		if h != "" {
			hosts = append(hosts, h)
		}
	}
	sort.Strings(hosts)

	// entries without host come last
	hosts = append(hosts, "")

	var routes []string
	for _, h := range hosts {
		for _, r := range t[h] {
			routes = append(routes, r.config(addWeight)...)
		}
	}
	return routes
}

// String returns the routing table as config file which can
// be read by Parse() again.
func (t Table) String() string {
	return strings.Join(t.Config(false), "\n")
}
