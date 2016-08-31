package route

import (
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"

	"github.com/eBay/fabio/metrics"
)

// Route maps a path prefix to one or more target URLs.
// routes can have a share value which describes the
// amount of traffic this route should get. You can specify
// that a route should get a fixed percentage of the traffic
// independent of how many instances are running.
type Route struct {
	// Host contains the host of the route.
	// not used for routing but for config generation
	// Table has a map with the host as key
	// for faster lookup and smaller search space.
	Host string

	// Path is the path prefix from a request uri
	Path string

	// Targets contains the list of URLs
	Targets []*Target

	// wTargets contains 100 targets distributed
	// according to their weight and ordered RR in the
	// same order as targets
	wTargets []*Target

	// total contains the total number of requests for this route.
	// Used by the RRPicker
	total uint64
}

func newRoute(host, path string) *Route {
	return &Route{Host: host, Path: path}
}

func (r *Route) addTarget(service string, targetURL *url.URL, fixedWeight float64, tags []string) {
	if fixedWeight < 0 {
		fixedWeight = 0
	}

	name, err := metrics.TargetName(service, r.Host, r.Path, targetURL)
	if err != nil {
		log.Printf("[ERROR] Invalid metrics name: %s", err)
		name = "unknown"
	}
	timer := ServiceRegistry.GetTimer(name)

	t := &Target{Service: service, Tags: tags, URL: targetURL, FixedWeight: fixedWeight, Timer: timer, timerName: name}
	r.Targets = append(r.Targets, t)
	r.weighTargets()
}

func (r *Route) delService(service string) {
	var clone []*Target
	for _, t := range r.Targets {
		if t.Service == service {
			continue
		}
		clone = append(clone, t)
	}
	r.Targets = clone
	r.weighTargets()
}

func (r *Route) delTarget(service string, targetURL *url.URL) {
	var clone []*Target
	for _, t := range r.Targets {
		if t.Service == service && t.URL.String() == targetURL.String() {
			continue
		}
		clone = append(clone, t)
	}
	r.Targets = clone
	r.weighTargets()
}

func (r *Route) setWeight(service string, weight float64, tags []string) int {
	loop := func(w float64) int {
		n := 0
		for _, t := range r.Targets {
			if service != "" && t.Service != service {
				continue
			}
			if len(tags) > 0 && !contains(t.Tags, tags) {
				continue
			}
			n++
			t.FixedWeight = w
		}
		return n
	}

	// if we have multiple matching targets
	// then we need to distribute the total
	// weight across all of them since the rule
	// states to assign only that percentage
	// of traffic to all matching routes combined.
	n := loop(0)
	w := weight / float64(n)
	loop(w)

	if n > 0 {
		r.weighTargets()
	}
	return n
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
func (r *Route) targetWeight(targetURL string) (n int) {
	for _, t := range r.wTargets {
		if t.URL.String() == targetURL {
			n++
		}
	}
	return n
}

func (r *Route) TargetConfig(t *Target, addWeight bool) string {
	s := fmt.Sprintf("route add %s %s %s", t.Service, r.Host+r.Path, t.URL)
	if addWeight {
		s += fmt.Sprintf(" weight %2.2f", t.Weight)
	} else if t.FixedWeight > 0 {
		s += fmt.Sprintf(" weight %.2f", t.FixedWeight)
	}
	if len(t.Tags) > 0 {
		s += fmt.Sprintf(" tags %q", strings.Join(t.Tags, ","))
	}
	return s
}

// config returns the route configuration in the config language.
// with the weights specified by the user.
func (r *Route) config(addWeight bool) []string {
	var cfg []string
	for _, t := range r.Targets {
		if t.Weight <= 0 {
			continue
		}
		cfg = append(cfg, r.TargetConfig(t, addWeight))
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
func (r *Route) weighTargets() {
	// how big is the fixed weighted traffic?
	var nFixed int
	var sumFixed float64
	for _, t := range r.Targets {
		if t.FixedWeight > 0 {
			nFixed++
			sumFixed += t.FixedWeight
		}
	}

	// normalize fixed weights up (sumFixed < 1) or down (sumFixed > 1)
	scale := 1.0
	if sumFixed > 1 || (nFixed == len(r.Targets) && sumFixed < 1) {
		scale = 1 / sumFixed
	}

	// compute the weight for the targets with dynamic weights
	dynamic := (1 - sumFixed) / float64(len(r.Targets)-nFixed)
	if dynamic < 0 {
		dynamic = 0
	}

	// assign the actual weight to each target
	for _, t := range r.Targets {
		if t.FixedWeight > 0 {
			t.Weight = t.FixedWeight * scale
		} else {
			t.Weight = dynamic
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

	slotCount := make(byN, len(r.Targets))
	for i, t := range r.Targets {
		slotCount[i].i = i
		slotCount[i].n = int(float64(wantSlots)*t.Weight + 0.5)
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
			slots[next] = r.Targets[c.i]
			next = (next + step) % gotSlots
		}
	}

	r.wTargets = slots
}

type byN []struct{ i, n int }

func (r byN) Len() int           { return len(r) }
func (r byN) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r byN) Less(i, j int) bool { return r[i].n < r[j].n }
