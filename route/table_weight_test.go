package route

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestWeight(t *testing.T) {
	tests := []struct {
		desc    string
		in, out func() []string
	}{
		{
			"dyn weight 1 -> auto distribution",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 1.0000`,
				}
			},
		},

		{
			"dyn weight 2 -> auto distribution",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/`,
					`route add svc /foo http://bar:222/`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 0.5000`,
					`route add svc /foo http://bar:222/ weight 0.5000`,
				}
			},
		},

		{
			"dyn weight 3 -> auto distribution",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/`,
					`route add svc /foo http://bar:222/`,
					`route add svc /foo http://bar:333/`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 0.3333`,
					`route add svc /foo http://bar:222/ weight 0.3333`,
					`route add svc /foo http://bar:333/ weight 0.3333`,
				}
			},
		},

		{
			"fixed weight 0 -> auto distribution",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 0`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 1.0000`,
				}
			},
		},

		{
			"only fixed weights and sum(fixedWeight) < 1 -> normalize to 100%",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 0.2`,
					`route add svc /foo http://bar:222/ weight 0.3`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 0.4000`,
					`route add svc /foo http://bar:222/ weight 0.6000`,
				}
			},
		},

		{
			"only fixed weights and sum(fixedWeight) > 1 -> normalize to 100%",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 2`,
					`route add svc /foo http://bar:222/ weight 3`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 0.4000`,
					`route add svc /foo http://bar:222/ weight 0.6000`,
				}
			},
		},

		{
			"multiple entries for same instance with no fixed weight -> de-duplication",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/`,
					`route add svc /foo http://bar:111/`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 1.0000`,
				}
			},
		},

		{
			"multiple entries with no fixed weight -> even distribution",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/`,
					`route add svc /foo http://bar:222/`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 0.5000`,
					`route add svc /foo http://bar:222/ weight 0.5000`,
				}
			},
		},

		{
			"multiple entries with de-dup and no fixed weight -> even distribution",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/`,
					`route add svc /foo http://bar:111/`,
					`route add svc /foo http://bar:222/`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 0.5000`,
					`route add svc /foo http://bar:222/ weight 0.5000`,
				}
			},
		},

		{
			"mixed fixed and auto weights -> even distribution of remaining weight across non-fixed weighted targets",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/`,
					`route add svc /foo http://bar:222/`,
					`route add svc /foo http://bar:333/ weight 0.5`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 0.2500`,
					`route add svc /foo http://bar:222/ weight 0.2500`,
					`route add svc /foo http://bar:333/ weight 0.5000`,
				}
			},
		},

		{
			"fixed weight == 100% -> route only to fixed weighted targets",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/`,
					`route add svc /foo http://bar:222/ weight 0.2500`,
					`route add svc /foo http://bar:333/ weight 0.7500`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:222/ weight 0.2500`,
					`route add svc /foo http://bar:333/ weight 0.7500`,
				}
			},
		},

		{
			"fixed weight > 100%  -> route only to fixed weighted targets and normalize weight",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/`,
					`route add svc /foo http://bar:222/ weight 1`,
					`route add svc /foo http://bar:333/ weight 3`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:222/ weight 0.2500`,
					`route add svc /foo http://bar:333/ weight 0.7500`,
				}
			},
		},

		{
			"dynamic weight matched on service name",
			func() []string {
				return []string{
					`route add svca /foo http://bar:111/`,
					`route add svcb /foo http://bar:222/`,
					`route add svcb /foo http://bar:333/`,
					`route weight svcb /foo weight 0.1`,
				}
			},
			func() []string {
				return []string{
					`route add svca /foo http://bar:111/ weight 0.9000`,
					`route add svcb /foo http://bar:222/ weight 0.0500`,
					`route add svcb /foo http://bar:333/ weight 0.0500`,
				}
			},
		},

		{
			"dynamic weight matched on service name and tags",
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ tags "a"`,
					`route add svc /foo http://bar:222/ tags "b"`,
					`route add svc /foo http://bar:333/ tags "b"`,
					`route weight svc /foo weight 0.1 tags "b"`,
				}
			},
			func() []string {
				return []string{
					`route add svc /foo http://bar:111/ weight 0.9000 tags "a"`,
					`route add svc /foo http://bar:222/ weight 0.0500 tags "b"`,
					`route add svc /foo http://bar:333/ weight 0.0500 tags "b"`,
				}
			},
		},

		{
			"dynamic weight matched on tags",
			func() []string {
				return []string{
					`route add svca /foo http://bar:111/ tags "a"`,
					`route add svcb /foo http://bar:222/ tags "b"`,
					`route add svcb /foo http://bar:333/ tags "b"`,
					`route weight /foo weight 0.1 tags "b"`,
				}
			},
			func() []string {
				return []string{
					`route add svca /foo http://bar:111/ weight 0.9000 tags "a"`,
					`route add svcb /foo http://bar:222/ weight 0.0500 tags "b"`,
					`route add svcb /foo http://bar:333/ weight 0.0500 tags "b"`,
				}
			},
		},

		{
			"more than 1000 routes",
			func() (a []string) {
				for i := 0; i < 2504; i++ {
					a = append(a, fmt.Sprintf(`route add svc /foo http://bar:%d/`, i))
				}
				return a
			},
			func() (a []string) {
				for i := 0; i < 2504; i++ {
					a = append(a, fmt.Sprintf(`route add svc /foo http://bar:%d/ weight 0.0004`, i))
				}
				return a
			},
		},

		{
			"more than 1000 routes with a fixed route target",
			func() (a []string) {
				for i := 0; i < 2504; i++ {
					a = append(a, fmt.Sprintf(`route add svc /foo http://bar:%d/`, i))
				}
				a = append(a, `route add svc /foo http://static:12345/ tags "a"`)
				a = append(a, `route weight svc /foo weight 0.2 tags "a"`)
				return a
			},
			func() (a []string) {
				for i := 0; i < 2504; i++ {
					a = append(a, fmt.Sprintf(`route add svc /foo http://bar:%d/ weight 0.0003`, i))
				}
				a = append(a, `route add svc /foo http://static:12345/ weight 0.2000 tags "a"`)
				return a
			},
		},
	}

	atof := func(s string) float64 {
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(err)
		}
		return n
	}

	for _, tt := range tests {
		tt := tt // capture loop var
		t.Run(tt.desc, func(t *testing.T) {
			in, out := tt.in(), tt.out()

			// parse the routes
			start := time.Now()
			tbl, err := ParseString(strings.Join(in, "\n"))
			if err != nil {
				t.Fatalf("got %v want nil", err)
			}
			t.Logf("parsing %d routes took %s seconds\n", len(in), time.Since(start))

			// compare the generated routes with the normalized weights
			if got, want := tbl.Config(true), out; !reflect.DeepEqual(got, want) {
				t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
			}

			// fetch the route
			r := tbl.route("", "/foo")
			if r == nil {
				t.Fatalf("got nil want route /foo")
			}

			// check that there are at least some slots
			if len(r.wTargets) == 0 {
				t.Fatalf("got 0 targets want some")
			}

			// count how often the 'url' from 'route add svc / <url>'
			// appears in the list of wTargets for all the URLs
			// from the routes to determine whether the actual
			// distribution of each target within the wTarget slice
			// matches what we expect

			// pre-generate the target urls for comparison as this
			// will otherwise slow the test down significantly
			targetURLs := make([]string, len(r.wTargets))
			for i, tg := range r.wTargets {
				targetURLs[i] = tg.URL.Scheme + "://" + tg.URL.Host + tg.URL.Path
			}

			for _, s := range out {
				// skip the 'route weight' lines
				if !strings.HasPrefix(s, "route add") {
					continue
				}

				// route add <svc> <path> <url> weight <weight> ...`,
				p := strings.Split(s, " ")
				svcurl, count := p[4], 0
				for _, tg := range targetURLs {
					if tg == svcurl {
						count++
					}
				}

				// calc the weight of the route as nSlots/totalSlots
				gotWeight := float64(count) / float64(len(r.wTargets))

				// round the weight down to the number of decimal points
				// supported by maxSlots
				gotWeight = float64(int(gotWeight*float64(maxSlots))) / float64(maxSlots)

				// we want the weight as specified in the generated config
				wantWeight := atof(p[6])

				// check that the actual weight is within 2% of the computed weight
				if math.Abs(gotWeight-wantWeight) > 0.02 {
					t.Errorf("got weight %f want %f", gotWeight, wantWeight)
				}

				// TODO(fs): verify distriibution of targets across the ring
			}
		})
	}
}
