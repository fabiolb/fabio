package config

import (
	"flag"
	"reflect"
	"testing"

	"github.com/magiconair/properties"
)

func TestParseFlags(t *testing.T) {
	props := func(s string) *properties.Properties {
		return properties.MustLoadString(s)
	}

	tests := []struct {
		desc   string
		args   []string
		env    []string
		prefix []string
		props  string
		a      []string
		kv     map[string]string
		kvs    []map[string]string
		v      string
	}{
		{
			desc:  "cmdline should win",
			args:  []string{"-v", "cmdline"},
			env:   []string{"v=env"},
			props: "v=props",
			v:     "cmdline",
		},
		{
			desc:  "env should win",
			env:   []string{"v=env"},
			props: "v=props",
			v:     "env",
		},
		{
			desc:   "env with prefix should win",
			env:    []string{"v=env", "p_v=prefix"},
			prefix: []string{"p_"},
			props:  "v=props",
			v:      "prefix",
		},
		{
			desc:  "props should win",
			props: "v=props",
			v:     "props",
		},
		{
			desc: "string slice in cmdline",
			args: []string{"-a", "1,2,3"},
			a:    []string{"1", "2", "3"},
		},
		{
			desc: "string slice in env",
			env:  []string{"a=1,2,3"},
			a:    []string{"1", "2", "3"},
		},
		{
			desc:  "string slice in props",
			props: "a=1,2,3",
			a:     []string{"1", "2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var a []string
			var v string
			f := NewFlagSet("test", flag.ExitOnError)
			f.StringVar(&v, "v", "", "")
			f.StringSliceVar(&a, "a", nil, "")
			err := f.ParseFlags(tt.args, tt.env, tt.prefix, props(tt.props))
			if err != nil {
				t.Errorf("got %v want nil", err)
			}
			if got, want := v, tt.v; got != want {
				t.Errorf("got %q want %q", got, want)
			}
			if got, want := a, tt.a; !reflect.DeepEqual(got, want) {
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}

func TestDefaults(t *testing.T) {
	a, aDefault := []string{}, []string{"x"}
	v, vDefault := "", "x"

	f := NewFlagSet("test", flag.ExitOnError)
	f.StringVar(&v, "v", vDefault, "")
	f.StringSliceVar(&a, "a", aDefault, "")

	if got, want := v, vDefault; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := a, aDefault; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}
