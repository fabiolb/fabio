package config

import (
	"flag"
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
	}

	for i, tt := range tests {
		var v string
		f := NewFlagSet("test", flag.ExitOnError)
		f.StringVar(&v, "v", "default", "")
		err := f.ParseFlags(tt.args, tt.env, tt.prefix, props(tt.props))
		if err != nil {
			t.Errorf("%d -%s: got %v want nil", i, tt.desc, err)
		}
		if got, want := v, tt.v; got != want {
			t.Errorf("%d - %s: got %q want %q", i, tt.desc, got, want)
		}
	}
}
