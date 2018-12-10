package route

import (
	"fmt"
	"log"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/fabiolb/fabio/metrics"
	"github.com/gobwas/glob"
)

// Route maps a path prefix to one or more target URLs.
// routes can have a weight value which describes the
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

	// wTargets contains targets distributed according to their weight
	wTargets []*Target

	// total contains the total number of requests for this route.
	// Used by the RRPicker
	total uint64

	// Glob represents compiled pattern.
	Glob glob.Glob
}

func (r *Route) addTarget(service string, targetURL *url.URL, fixedWeight float64, tags []string, opts map[string]string) {
	if fixedWeight < 0 {
		fixedWeight = 0
	}

	// de-dup existing target
	for _, t := range r.Targets {
		if t.Service == service && t.URL.String() == targetURL.String() && t.FixedWeight == fixedWeight && reflect.DeepEqual(t.Tags, tags) {
			return
		}
	}

	name, err := metrics.TargetName(service, r.Host, r.Path, targetURL)
	if err != nil {
		log.Printf("[ERROR] Invalid metrics name: %s", err)
		name = "unknown"
	}

	t := &Target{
		Service:     service,
		Tags:        tags,
		Opts:        opts,
		URL:         targetURL,
		FixedWeight: fixedWeight,
		Timer:       ServiceRegistry.GetTimer(name),
		TimerName:   name,
	}

	if opts != nil {
		t.StripPath = opts["strip"]
		t.TLSSkipVerify = opts["tlsskipverify"] == "true"
		t.Host = opts["host"]

		if opts["redirect"] != "" {
			t.RedirectCode, err = strconv.Atoi(opts["redirect"])
			if err != nil {
				log.Printf("[ERROR] redirect status code should be numeric in 3xx range. Got: %s", opts["redirect"])
			} else if t.RedirectCode < 300 || t.RedirectCode > 399 {
				t.RedirectCode = 0
				log.Printf("[ERROR] redirect status code should be in 3xx range. Got: %s", opts["redirect"])
			}
		}

		if err = t.ProcessAccessRules(); err != nil {
			log.Printf("[ERROR] failed to process access rules: %s",
				err.Error())
		}

		t.AuthScheme = opts["auth"]
	}

	r.Targets = append(r.Targets, t)
	r.weighTargets()
}

func (r *Route) filter(skip func(t *Target) bool) {
	var clone []*Target
	for _, t := range r.Targets {
		if skip(t) {
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

func (r *Route) TargetConfig(t *Target, addWeight bool) string {
	s := fmt.Sprintf("route add %s %s %s", t.Service, r.Host+r.Path, t.URL)
	if addWeight {
		s += fmt.Sprintf(" weight %2.4f", t.Weight)
	} else if t.FixedWeight > 0 {
		s += fmt.Sprintf(" weight %.4f", t.FixedWeight)
	}
	if len(t.Tags) > 0 {
		s += fmt.Sprintf(" tags %q", strings.Join(t.Tags, ","))
	}
	if len(t.Opts) > 0 {
		var keys []string
		for k := range t.Opts {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var vals []string
		for _, k := range keys {
			vals = append(vals, k+"="+t.Opts[k])
		}
		s += fmt.Sprintf(" opts \"%s\"", strings.Join(vals, " "))
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

// maxSlots defines the maximum number of slots on the ring for
// weighted round-robin distribution for a single route. Consequently,
// this then defines the maximum number of separate instances that can
// serve a single route. maxSlots must be a power of ten.
const maxSlots = 1e4 // 10000

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

	// if there are no targets with fixed weight then each target simply gets
	// an equal amount of traffic
	if nFixed == 0 {
		w := 1.0 / float64(len(r.Targets))
		for _, t := range r.Targets {
			t.Weight = w
		}
		r.wTargets = r.Targets
		return
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

	// distribute the targets on a ring suitable for weighted round-robin
	// distribution
	//
	// This is done in two steps:
	//
	// Step one determines the necessary ring size to distribute the targets
	// according to their weight with reasonable accuracy. For example, two
	// targets with 50% weight fit in a ring of size 2 whereas two targets with
	// 10% and 90% weight require a ring of size 10.
	//
	// To keep it simple we allocate 10000 slots which provides slots to all
	// targets with at least a weight of 0.01%. In addition, we guarantee that
	// every target with a weight > 0 gets at least one slot. The case where
	// all targets get an equal share of traffic is handled earlier so this is
	// for situations with some fixed weight.
	//
	// Step two distributes the targets onto the ring spacing them out evenly
	// so that iterating over the ring performs the weighted rr distribution.
	// For example, a 50/50 distribution on a ring of size 10 should be
	// 0101010101 instead of 0000011111.
	//
	// To ensure that targets with smaller weights are properly placed we place
	// them on the ring first by sorting the targets by slot count.
	//
	// TODO(fs): I assume that this is some sort of mathematical problem
	// (coloring, optimizing, ...) but I don't know which. Happy to make this
	// more formal, if possible.
	//
	slots := make(byN, len(r.Targets))
	usedSlots := 0
	for i, t := range r.Targets {
		n := int(float64(maxSlots) * t.Weight)
		if n == 0 && t.Weight > 0 {
			n = 1
		}
		slots[i].i = i
		slots[i].n = n
		usedSlots += n
	}

	sort.Sort(slots)
	targets := make([]*Target, usedSlots)
	for _, s := range slots {
		if s.n <= 0 {
			continue
		}

		next, step := 0, usedSlots/s.n
		for k := 0; k < s.n; k++ {
			// find the next empty slot
			for targets[next] != nil {
				next = (next + 1) % usedSlots
			}

			// use slot and move to next one
			targets[next] = r.Targets[s.i]
			next = (next + step) % usedSlots
		}
	}

	r.wTargets = targets
}

type byN []struct{ i, n int }

func (r byN) Len() int           { return len(r) }
func (r byN) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r byN) Less(i, j int) bool { return r[i].n < r[j].n }
