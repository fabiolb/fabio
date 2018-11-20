package route

import (
	"reflect"
	"regexp"
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
			desc: "RouteAddTCPService",
			in:   `route add svc :1234 tcp://1.2.3.4:5678`,
			out:  []*RouteDef{{Cmd: RouteAddCmd, Service: "svc", Src: ":1234", Dst: "tcp://1.2.3.4:5678"}},
		},
		{
			desc: "RouteAddGRPCService",
			in:   `route add svc :1234 grpc://1.2.3.4:5678`,
			out:  []*RouteDef{{Cmd: RouteAddCmd, Service: "svc", Src: ":1234", Dst: "grpc://1.2.3.4:5678"}},
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
			desc: "RouteDelTags",
			in:   `route del tags "a,b"`,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Tags: []string{"a", "b"}}},
		},
		{
			desc: "RouteDelTagsMoreSpaces",
			in:   `route  del  tags  " a , b "`,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Tags: []string{"a", "b"}}},
		},
		{
			desc: "RouteDelService",
			in:   `route del svc`,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Service: "svc"}},
		},
		{
			desc: "RouteDelServiceTags",
			in:   `route del svc tags "a,b"`,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Service: "svc", Tags: []string{"a", "b"}}},
		},
		{
			desc: "RouteDelServiceTagsMoreSpaces",
			in:   `route  del  svc  tags  " a , b "`,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Service: "svc", Tags: []string{"a", "b"}}},
		},
		{
			desc: "RouteDelServiceSrc",
			in:   `route del svc /prefix`,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Service: "svc", Src: "/prefix"}},
		},
		{
			desc: "RouteDelTCPServiceSrc",
			in:   `route del svc :1234`,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Service: "svc", Src: ":1234"}},
		},
		{
			desc: "RouteDelServiceSrcDst",
			in:   `route del svc /prefix http://1.2.3.4/`,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Service: "svc", Src: "/prefix", Dst: "http://1.2.3.4/"}},
		},
		{
			desc: "RouteDelTCPServiceSrcDst",
			in:   `route del svc :1234 tcp://1.2.3.4:5678`,
			out:  []*RouteDef{{Cmd: RouteDelCmd, Service: "svc", Src: ":1234", Dst: "tcp://1.2.3.4:5678"}},
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

	deref := func(def []*RouteDef) (defs []RouteDef) {
		for _, d := range def {
			defs = append(defs, *d)
		}
		return
	}

	run := func(in string, def []*RouteDef, fail bool, parseFn func(string) ([]*RouteDef, error)) {
		out, err := parseFn(in)
		switch {
		case err == nil && fail:
			t.Errorf("got error nil want fail")
			return
		case err != nil && !fail:
			t.Errorf("got error %v want nil", err)
			return
		case err != nil:
			if !reSyntaxError.MatchString(err.Error()) {
				t.Errorf("got error %q want 'syntax error.*'", err)
			}
			return
		}
		if got, want := out, def; !reflect.DeepEqual(got, want) {
			t.Errorf("\ngot  %#v\nwant %#v", deref(got), deref(want))
		}
	}

	for _, tt := range tests {
		t.Run("Parse-"+tt.desc, func(t *testing.T) { run(tt.in, tt.out, tt.fail, Parse) })
	}
}

func TestParseAliases(t *testing.T) {
	tests := []struct {
		desc string
		in   string
		out  []string
		fail bool
	}{
		// error flows
		{"FailEmpty", ``, nil, false},
		{"FailNoRoute", `bang`, nil, true},
		{"FailRouteNoCmd", `route x`, nil, true},
		{"FailRouteAddNoService", `route add`, nil, true},
		{"FailRouteAddNoSrc", `route add svc`, nil, true},

		// happy flows with and without aliases
		{
			desc: "RouteAddServiceWithoutAlias",
			in:   `route add alpha-be alpha/ http://1.2.3.4/ opts "strip=/path proto=https"`,
			out:  []string(nil),
		},
		{
			desc: "RouteAddServiceWithAlias",
			in:   `route add alpha-be alpha/ http://1.2.3.4/ opts "strip=/path proto=https register=alpha"`,
			out:  []string{"alpha"},
		},
		{
			desc: "RouteAddServicesWithoutAliases",
			in: `route add alpha-be alpha/ http://1.2.3.4/ opts "strip=/path proto=tcp"
			route add bravo-be bravo/ http://1.2.3.5/
			route add charlie-be charlie/ http://1.2.3.6/ opts "host=charlie"`,
			out: []string(nil),
		},
		{
			desc: "RouteAddServicesWithAliases",
			in: `route add alpha-be alpha/ http://1.2.3.4/ opts "register=alpha"
			route add bravo-be bravo/ http://1.2.3.5/ opts "strip=/foo register=bravo"
			route add charlie-be charlie/ http://1.2.3.5/ opts "host=charlie proto=https"
			route add delta-be delta/ http://1.2.3.5/ opts "host=delta proto=https register=delta"`,
			out: []string{"alpha", "bravo", "delta"},
		},
	}

	reSyntaxError := regexp.MustCompile(`syntax error`)

	run := func(in string, aliases []string, fail bool, parseFn func(string) ([]string, error)) {
		out, err := parseFn(in)
		switch {
		case err == nil && fail:
			t.Errorf("got error nil want fail")
			return
		case err != nil && !fail:
			t.Errorf("got error %v want nil", err)
			return
		case err != nil:
			if !reSyntaxError.MatchString(err.Error()) {
				t.Errorf("got error %q want 'syntax error.*'", err)
			}
			return
		}
		if got, want := out, aliases; !reflect.DeepEqual(got, want) {
			t.Errorf("\ngot  %#v\nwant %#v", got, want)
		}
	}

	for _, tt := range tests {
		t.Run("ParseAliases-"+tt.desc, func(t *testing.T) { run(tt.in, tt.out, tt.fail, ParseAliases) })
	}
}
