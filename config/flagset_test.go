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
		{
			desc: "kv in cmdline",
			args: []string{"-kv", "a=1;b=2"},
			kv:   map[string]string{"a": "1", "b": "2"},
		},
		{
			desc: "kv in env",
			env:  []string{"kv=a=1;b=2"},
			kv:   map[string]string{"a": "1", "b": "2"},
		},
		{
			desc:  "kv in props",
			props: "kv=a=1;b=2",
			kv:    map[string]string{"a": "1", "b": "2"},
		},
		{
			desc: "kv slice in cmdline",
			args: []string{"-kvs", "a=1;b=2,c=3;d=4"},
			kvs:  []map[string]string{{"a": "1", "b": "2"}, {"c": "3", "d": "4"}},
		},
		{
			desc: "kv slice in env",
			env:  []string{"kvs=a=1;b=2,c=3;d=4"},
			kvs:  []map[string]string{{"a": "1", "b": "2"}, {"c": "3", "d": "4"}},
		},
		{
			desc:  "kv slice in props",
			props: "kvs=a=1;b=2,c=3;d=4",
			kvs:   []map[string]string{{"a": "1", "b": "2"}, {"c": "3", "d": "4"}},
		},
		{
			desc:  "kv slice with spaces",
			props: "kvs= a = 1 ; b = 2 , c = 3 ; d = 4 ",
			kvs:   []map[string]string{{"a": " 1 ", "b": " 2 "}, {"c": " 3 ", "d": " 4 "}},
		},
	}

	for i, tt := range tests {
		var a []string
		var kv map[string]string
		var kvs []map[string]string
		var v string
		f := NewFlagSet("test", flag.ExitOnError)
		f.StringVar(&v, "v", "", "")
		f.KVVar(&kv, "kv", nil, "")
		f.KVSliceVar(&kvs, "kvs", nil, "")
		f.StringSliceVar(&a, "a", nil, "")
		err := f.ParseFlags(tt.args, tt.env, tt.prefix, props(tt.props))
		if err != nil {
			t.Errorf("%d -%s: got %v want nil", i, tt.desc, err)
		}
		if got, want := v, tt.v; got != want {
			t.Errorf("%d - %s: got %q want %q", i, tt.desc, got, want)
		}
		if got, want := kv, tt.kv; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: %s: got %v want %v", i, tt.desc, got, want)
		}
		if got, want := kvs, tt.kvs; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: %s: got %v want %v", i, tt.desc, got, want)
		}
		if got, want := a, tt.a; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: %s: got %v want %v", i, tt.desc, got, want)
		}
	}
}

func TestDefaults(t *testing.T) {
	a, aDefault := []string{}, []string{"x"}
	kv, kvDefault := map[string]string{}, map[string]string{"x": "y"}
	kvs, kvsDefault := []map[string]string{}, []map[string]string{{"x": "y"}}
	v, vDefault := "", "x"

	f := NewFlagSet("test", flag.ExitOnError)
	f.StringVar(&v, "v", vDefault, "")
	f.KVVar(&kv, "kv", kvDefault, "")
	f.KVSliceVar(&kvs, "kvs", kvsDefault, "")
	f.StringSliceVar(&a, "a", aDefault, "")

	if got, want := v, vDefault; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := kv, kvDefault; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := kvs, kvsDefault; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := a, aDefault; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}
