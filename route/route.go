package route

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/eBay/fabio/metrics"

	gometrics "github.com/eBay/fabio/_third_party/github.com/rcrowley/go-metrics"
)

// route maps a path prefix to one or more target URLs.
// routes can have a share value which describes the
// amount of traffic this route should get. You can specify
// that a route should get a fixed percentage of the traffic
// independent of how many instances are running.
type route struct {
	// host contains the host of the route.
	// not used for routing but for config generation
	// Table has a map with the host as key
	// for faster lookup and smaller search space.
	host string

	// path is the path prefix from a request uri
	path string

	// targets contains the list of URLs
	targets []*Target

	// wTargets contains 100 targets distributed
	// according to their weight and ordered RR in the
	// same order as targets
	wTargets []*Target

	// total contains the total number of requests for this route.
	// Used by the RRPicker
	total uint64
}

type Target struct {
	// service is the name of the service the targetURL points to
	service string

	// tags are the list of tags for this target
	tags []string

	// URL is the endpoint the service instance listens on
	URL *url.URL

	// fixedWeight is the weight assigned to this target.
	// If the value is 0 the targets weight is dynamic.
	fixedWeight float64

	// weight is the actual weight for this service in percent.
	weight float64

	// timer measures throughput and latency of this target
	Timer gometrics.Timer
}

func newRoute(host, path string) *route {
	return &route{host: host, path: path}
}

func (r *route) addTarget(service string, targetURL *url.URL, fixedWeight float64, tags []string) {
	if fixedWeight < 0 {
		fixedWeight = 0
	}

	name := metrics.TargetName(service, r.host, r.path, targetURL)
	timer := gometrics.GetOrRegisterTimer(name, gometrics.DefaultRegistry)

	t := &Target{service: service, tags: tags, URL: targetURL, fixedWeight: fixedWeight, Timer: timer}
	r.targets = append(r.targets, t)
	r.weighTargets()
}

func (r *route) delService(service string) {
	var clone []*Target
	for _, t := range r.targets {
		if t.service == service {
			continue
		}
		clone = append(clone, t)
	}
	r.targets = clone
	r.weighTargets()
}

func (r *route) delTarget(service string, targetURL *url.URL) {
	var clone []*Target
	for _, t := range r.targets {
		if t.service == service && t.URL.String() == targetURL.String() {
			continue
		}
		clone = append(clone, t)
	}
	r.targets = clone
	r.weighTargets()
}

func (r *route) setWeight(weight float64, tags []string) int {
	updated := 0
	for _, t := range r.targets {
		if contains(t.tags, tags) {
			t.fixedWeight = weight
			updated++
		}
	}
	if updated > 0 {
		r.weighTargets()
	}
	return updated
}

func contains(src, dst []string) bool {
	for _, d := range dst {
		found := false
		for _, s := range src {
			if s == d {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// targetWeight returns how often target is in wTargets.
func (r *route) targetWeight(targetURL string) (n int) {
	for _, t := range r.wTargets {
		if t.URL.String() == targetURL {
			n++
		}
	}
	return n
}

// config returns the route configuration in the config language.
// with the weights specified by the user.
func (r *route) config(addWeight bool) []string {
	var cfg []string
	for _, t := range r.targets {
		if t.weight <= 0 {
			continue
		}

		s := fmt.Sprintf("route add %s %s %s", t.service, r.host+r.path, t.URL)
		if addWeight {
			s += fmt.Sprintf(" weight %2.2f", t.weight)
		} else if t.fixedWeight > 0 {
			s += fmt.Sprintf(" weight %.2f", t.fixedWeight)
		}
		if len(t.tags) > 0 {
			s += fmt.Sprintf(" tags %q", strings.Join(t.tags, ","))
		}
		cfg = append(cfg, s)
	}
	return cfg
}

// weighTargets computes the share of traffic each target receives based
// on its weight and the weight of the other targets.
//
// Traffic is first distributed to targets with a fixed weight. If the sum of
// all fixed weights exceeds 100% then they are normalized to 100%.
//
// Targets with a dynamic weight will receive an equal share of the remaining
// traffic if there is any left.
func (r *route) weighTargets() {
	// how big is the fixed weighted traffic?
	var nFixed int
	var sumFixed float64
	for _, t := range r.targets {
		if t.fixedWeight > 0 {
			nFixed++
			sumFixed += t.fixedWeight
		}
	}

	// normalize fixed weights up (sumFixed < 1) or down (sumFixed > 1)
	scale := 1.0
	if sumFixed > 1 || (nFixed == len(r.targets) && sumFixed < 1) {
		scale = 1 / sumFixed
	}

	// compute the weight for the targets with dynamic weights
	dynamic := (1 - sumFixed) / float64(len(r.targets)-nFixed)
	if dynamic < 0 {
		dynamic = 0
	}

	// assign the actual weight to each target
	for _, t := range r.targets {
		if t.fixedWeight > 0 {
			t.weight = t.fixedWeight * scale
		} else {
			t.weight = dynamic
		}
	}

	// Distribute the targets on a ring with N slots. The distance
	// between two entries for the same target should be N/count slots
	// apart to achieve even distribution. count is the number of slots the
	// target should get based on its weight.
	// To achieve this we first determine count per target and then sort that
	// from smallest to largest to distribute the targets with lesser weight
	// more evenly. For that we pick a random starting point on the ring and
	// move clockwise until we find a free spot. The the next slot is N/count
	// slots away. If it is occupied we again move clockwise until we find
	// a free slot.

	// number of slots we want to use and number of slots we will actually use
	// because of rounding errors
	gotSlots, wantSlots := 0, 100

	slotCount := make(byN, len(r.targets))
	for i, t := range r.targets {
		slotCount[i].i = i
		slotCount[i].n = int(float64(wantSlots)*t.weight + 0.5)
		gotSlots += slotCount[i].n
	}
	sort.Sort(slotCount)

	slots := make([]*Target, gotSlots)
	for _, c := range slotCount {
		if c.n <= 0 {
			continue
		}

		next, step := 0, gotSlots/c.n
		for k := 0; k < c.n; k++ {
			// find the next empty slot
			for slots[next] != nil {
				next = (next + 1) % gotSlots
			}

			// use slot and move to next one
			slots[next] = r.targets[c.i]
			next = (next + step) % gotSlots
		}
	}

	r.wTargets = slots
}

type byN []struct{ i, n int }

func (r byN) Len() int           { return len(r) }
func (r byN) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r byN) Less(i, j int) bool { return r[i].n < r[j].n }
