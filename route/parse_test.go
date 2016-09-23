package route

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		desc string
		in   string
		out  []*RouteDef
		fail bool
	}{
		// error flows
		{"FailEmpty", ``, nil, false},
		{"FailNoRoute", `bang`, nil, true},
		{"FailRouteNoCmd", `route x`, nil, true},
		{"FailRouteAddNoService", `route add`, nil, true},
		{"FailRouteAddNoSrc", `route add svc`, nil, true},

		// happy flows
		{
			desc: "RouteAddService",
			in:   `route add svc /prefix http://1.2.3.4/`,
			out:  []*RouteDef{{Cmd: RouteAddCmd, Service: "svc", Src: "/prefix", Dst: "http://1.2.3.4/"}},
		},
		{
			desc: "RouteAddServiceWeight",
			in:   `route add svc /prefix http://1.2.3.4/ weight 1.2`,
			out:  []*RouteDef{{Cmd: RouteAddCmd, Service: "svc", Src: "/prefix", Dst: "http://1.2.3.4/", Weight: 1.2}},
		},
		{
			desc: "RouteAddServiceWeightTags",
			in:   `route add svc /prefix http://1.2.3.4/ weight 1.2 tags "a,b"`,
			out:  []*RouteDef{{Cmd: RouteAddCmd, Service: "svc", Src: "/prefix", Dst: "http://1.2.3.4/", Weight: 1.2, Tags: []string{"a", "b"}}},
		},
		{
			desc: "RouteAddServiceWeightOpts",
			in:   `route add svc /prefix http://1.2.3.4/ weight 1.2 opts "foo=bar baz=bang blimp"`,
			out:  []*RouteDef{{Cmd: RouteAddCmd, Service: "svc", Src: "/prefix", Dst: "http://1.2.3.4/", Weight: 1.2, Opts: map[string]string{"foo": "bar", "baz": "bang", "blimp": ""}}},
		},
		{
			desc: "RouteAddServiceWeightTagsOpts",
			in:   `route add svc /prefix http://1.2.3.4/ weight 1.2 tags "a,b" opts "foo=bar baz=bang blimp"`,
			out:  []*RouteDef{{Cmd: RouteAddCmd, Service: "svc", Src: "/prefix", Dst: "http://1.2.3.4/", Weight: 1.2, Tags: []string{"a", "b"}, Opts: map[string]string{"foo": "bar", "baz": "bang", "blimp": ""}}},
		},
		{
			desc: "RouteAddServiceWeightTagsOptsMoreSpaces",
			in:   ` route  add  svc  /prefix  http://1.2.3.4/  weight  1.2  tags  " a , b "  opts  "foo=bar  baz=bang  blimp" `,
			out:  []*RouteDef{{Cmd: RouteAddCmd, Service: "svc", Src: "/prefix", Dst: "http://1.2.3.4/", Weight: 1.2, Tags: []string{"a", "b"}, Opts: map[string]string{"foo": "bar", "baz": "bang", "blimp": ""}}},
		},
		{
			desc: "RouteAddTags",
			in:   `route add svc /prefix http://1.2.3.4/ tags "a,b"`,
			out:  []*RouteDef{{Cmd: RouteAddCmd, Service: "svc", Src: "/prefix", Dst: "http://1.2.3.4/", Tags: []string{"a", "b"}}},
		},
		{
			desc: "RouteAddTagsOpts",
			in:   `route add svc /prefix http://1.2.3.4/ tags "a,b" opts "foo=bar baz=bang blimp"`,
			out:  []*RouteDef{{Cmd: RouteAddCmd, Service: "svc", Src: "/prefix", Dst: "http://1.2.3.4/", Tags: []string{"a", "b"}, Opts: map[string]string{"foo": "bar", "baz": "bang", "blimp": ""}}},
		},
		{
			desc: "RouteAddOpts",
			in:   `route add svc /prefix http://1.2.3.4/ opts "foo=bar baz=bang blimp"`,
			out:  []*RouteDef{{Cmd: RouteAddCmd, Service: "svc", Src: "/prefix", Dst: "http://1.2.3.4/", Opts: map[string]string{"foo": "bar", "baz": "bang", "blimp": ""}}},
		},
		{
			desc: "RouteDelService",
			in:   `route del svc`,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Service: "svc"}},
		},
		{
			desc: "RouteDelServiceSrc",
			in:   `route del svc /prefix`,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Service: "svc", Src: "/prefix"}},
		},
		{
			desc: "RouteDelServiceSrcDst",
			in:   `route del svc /prefix http://1.2.3.4/`,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Service: "svc", Src: "/prefix", Dst: "http://1.2.3.4/"}},
		},
		{
			desc: "RouteDelServiceSrcDstMoreSpaces",
			in:   ` route  del  svc  /prefix  http://1.2.3.4/ `,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Service: "svc", Src: "/prefix", Dst: "http://1.2.3.4/"}},
		},
		{
			desc: "RouteWeightServiceSrc",
			in:   `route weight svc /prefix weight 1.2`,
			out:  []*RouteDef{{Cmd: RouteWeightCmd, Service: "svc", Src: "/prefix", Weight: 1.2}},
		},
		{
			desc: "RouteWeightServiceSrcTags",
			in:   `route weight svc /prefix weight 1.2 tags "a,b"`,
			out:  []*RouteDef{{Cmd: RouteWeightCmd, Service: "svc", Src: "/prefix", Weight: 1.2, Tags: []string{"a", "b"}}},
		},
		{
			desc: "RouteWeightServiceSrcTagsMoreSpaces",
			in:   ` route  weight  svc  /prefix  weight  1.2  tags  " a , b " `,
			out:  []*RouteDef{{Cmd: RouteWeightCmd, Service: "svc", Src: "/prefix", Weight: 1.2, Tags: []string{"a", "b"}}},
		},
		{
			desc: "RouteWeightSrcTags",
			in:   `route weight /prefix weight 1.2 tags "a,b"`,
			out:  []*RouteDef{{Cmd: RouteWeightCmd, Src: "/prefix", Weight: 1.2, Tags: []string{"a", "b"}}},
		},
		{
			desc: "RouteWeightSrcTagsMoreSpaces",
			in:   ` route  weight  /prefix  weight  1.2  tags  " a , b " `,
			out:  []*RouteDef{{Cmd: RouteWeightCmd, Src: "/prefix", Weight: 1.2, Tags: []string{"a", "b"}}},
		},
	}

	reSyntaxError := regexp.MustCompile(`syntax error`)

	run := func(in string, def []*RouteDef, fail bool, parseFn func(string) ([]*RouteDef, error)) {
		out, err := parseFn(in)
		switch {
		case err == nil && fail:
			t.Errorf("got nil want fail")
		case err != nil && !fail:
			t.Errorf("got %v want nil", err)
			return
		case err != nil:
			if !reSyntaxError.MatchString(err.Error()) {
				t.Errorf("got %q want 'syntax error.*'", err)
			}
			return
		}
		if got, want := out, def; !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v want %+v", got, want)
		}
	}

	for _, tt := range tests {
		if !strings.Contains(tt.desc, "Opts") {
			t.Run("Parse-"+tt.desc, func(t *testing.T) { run(tt.in, tt.out, tt.fail, Parse) })
		}
		t.Run("ParseNew-"+tt.desc, func(t *testing.T) { run(tt.in, tt.out, tt.fail, ParseNew) })
	}
}
