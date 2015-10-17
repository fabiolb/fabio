package route

import (
	"reflect"
	"strings"
	"testing"
)

func TestWeight(t *testing.T) {
	tests := []struct {
		in, out []string
		counts  []int
	}{
		{ // no fixed weight -> auto distribution
			[]string{
				`route add svc /foo http://bar:111/`,
			},
			[]string{
				`route add svc /foo http://bar:111/ weight 1.00`,
			},
			[]int{100},
		},

		{ // fixed weight 0 -> auto distribution
			[]string{
				`route add svc /foo http://bar:111/ weight 0`,
			},
			[]string{
				`route add svc /foo http://bar:111/ weight 1.00`,
			},
			[]int{100},
		},

		{ // only fixed weights and sum(fixedWeight) < 1 -> normalize to 100%
			[]string{
				`route add svc /foo http://bar:111/ weight 0.2`,
				`route add svc /foo http://bar:222/ weight 0.3`,
			},
			[]string{
				`route add svc /foo http://bar:111/ weight 0.40`,
				`route add svc /foo http://bar:222/ weight 0.60`,
			},
			[]int{40, 60},
		},

		{ // only fixed weights and sum(fixedWeight) > 1 -> normalize to 100%
			[]string{
				`route add svc /foo http://bar:111/ weight 2`,
				`route add svc /foo http://bar:222/ weight 3`,
			},
			[]string{
				`route add svc /foo http://bar:111/ weight 0.40`,
				`route add svc /foo http://bar:222/ weight 0.60`,
			},
			[]int{40, 60},
		},

		// TODO(fs): should Table de-duplicate?
		{ // multiple entries with no fixed weight -> even distribution (same service)
			[]string{
				`route add svc /foo http://bar:111/`,
				`route add svc /foo http://bar:111/`,
			},
			[]string{
				`route add svc /foo http://bar:111/ weight 0.50`,
				`route add svc /foo http://bar:111/ weight 0.50`,
			},
			[]int{100, 100},
		},

		{ // multiple entries with no fixed weight -> even distribution
			[]string{
				`route add svc /foo http://bar:111/`,
				`route add svc /foo http://bar:222/`,
			},
			[]string{
				`route add svc /foo http://bar:111/ weight 0.50`,
				`route add svc /foo http://bar:222/ weight 0.50`,
			},
			[]int{50, 50},
		},

		{ // mixed fixed and auto weights -> even distribution of remaining weight across non-fixed weighted targets
			[]string{
				`route add svc /foo http://bar:111/`,
				`route add svc /foo http://bar:222/`,
				`route add svc /foo http://bar:333/ weight 0.5`,
			},
			[]string{
				`route add svc /foo http://bar:111/ weight 0.25`,
				`route add svc /foo http://bar:222/ weight 0.25`,
				`route add svc /foo http://bar:333/ weight 0.50`,
			},
			[]int{25, 25, 50},
		},

		{ // fixed weight == 100% -> route only to fixed weighted targets
			[]string{
				`route add svc /foo http://bar:111/`,
				`route add svc /foo http://bar:222/ weight 0.25`,
				`route add svc /foo http://bar:333/ weight 0.75`,
			},
			[]string{
				`route add svc /foo http://bar:222/ weight 0.25`,
				`route add svc /foo http://bar:333/ weight 0.75`,
			},
			[]int{0, 25, 75},
		},

		{ // fixed weight > 100%  -> route only to fixed weighted targets and normalize weight
			[]string{
				`route add svc /foo http://bar:111/`,
				`route add svc /foo http://bar:222/ weight 1`,
				`route add svc /foo http://bar:333/ weight 3`,
			},
			[]string{
				`route add svc /foo http://bar:222/ weight 0.25`,
				`route add svc /foo http://bar:333/ weight 0.75`,
			},
			[]int{0, 25, 75},
		},

		{ // fixed weight > 100%  -> route only to fixed weighted targets and normalize weight
			[]string{
				`route add svc /foo http://bar:111/ tags "a"`,
				`route add svc /foo http://bar:222/ tags "b"`,
				`route weight svc /foo weight 0.1 tags "b"`,
			},
			[]string{
				`route add svc /foo http://bar:111/ weight 0.90 tags "a"`,
				`route add svc /foo http://bar:222/ weight 0.10 tags "b"`,
			},
			[]int{90, 10},
		},
	}

	for i, tt := range tests {
		tbl, err := ParseString(strings.Join(tt.in, "\n"))
		if err != nil {
			t.Fatalf("%d: got %v want nil", i, err)
		}
		if got, want := tbl.Config(true), tt.out; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got\n%s\nwant\n%s", i, strings.Join(got, "\n"), strings.Join(want, "\n"))
		}

		// count url occurrences
		r := tbl.route("", "/foo")
		if r == nil {
			t.Fatalf("%d: got nil want route /foo", i)
		}
		for j, s := range tt.in {
			if !strings.HasPrefix(s, "route add") {
				continue
			}
			p := strings.Split(s, " ")
			if got, want := r.targetWeight(p[4]), tt.counts[j]; got != want {
				t.Errorf("%d: %s: got %d want %d", i, p[4], got, want)
			}
		}
	}
}
