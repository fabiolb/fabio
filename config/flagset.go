package config

import (
	"flag"
	"sort"
	"strings"

	"github.com/magiconair/properties"
)

// -- kvValue
type kvValue map[string]string

func newKVValue(val map[string]string, p *map[string]string) *kvValue {
	*p = val
	return (*kvValue)(p)
}

func kvParse(s string) kvValue {
	m := map[string]string{}
	for _, s := range strings.Split(s, ";") {
		p := strings.SplitN(s, "=", 2)
		if len(p) == 1 {
			m[p[0]] = ""
		} else {
			m[p[0]] = p[1]
		}
	}
	return m
}

func kvString(kv kvValue) string {
	var keys []string
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var p []string
	for _, k := range keys {
		p = append(p, k+"="+kv[k])
	}
	return strings.Join(p, ";")
}

func (v *kvValue) Set(s string) error {
	*v = kvParse(s)
	return nil
}

func (v *kvValue) Get() interface{} { return map[string]string(*v) }
func (v *kvValue) String() string   { return kvString(*v) }

// -- kvSliceValue
type kvSliceValue []map[string]string

func newKVSliceValue(val []map[string]string, p *[]map[string]string) *kvSliceValue {
	*p = val
	return (*kvSliceValue)(p)
}

func (v *kvSliceValue) Set(s string) error {
	*v = []map[string]string{}
	for _, x := range strings.Split(s, ",") {
		*v = append(*v, kvParse(x))
	}
	return nil
}

func (v *kvSliceValue) Get() interface{} { return []map[string]string(*v) }
func (v *kvSliceValue) String() string {
	var p []string
	for i := range *v {
		p = append(p, kvString((*v)[i]))
	}
	return strings.Join(p, ",")
}

// -- stringSliceValue
type stringSliceValue []string

func newStringSliceValue(val []string, p *[]string) *stringSliceValue {
	*p = val
	return (*stringSliceValue)(p)
}

func (v *stringSliceValue) Set(s string) error {
	*v = []string{}
	for _, x := range strings.Split(s, ",") {
		x = strings.TrimSpace(x)
		if x == "" {
			continue
		}
		*v = append(*v, x)
	}
	return nil
}

func (v *stringSliceValue) Get() interface{} { return []string(*v) }
func (v *stringSliceValue) String() string   { return strings.Join(*v, ",") }

// -- FlagSet
type FlagSet struct {
	flag.FlagSet
	set map[string]bool
}

func NewFlagSet(name string, errorHandling flag.ErrorHandling) *FlagSet {
	fs := &FlagSet{set: make(map[string]bool)}
	fs.Init(name, errorHandling)
	return fs
}

// IsSet returns true if a variable was set via any mechanism.
func (f *FlagSet) IsSet(name string) bool {
	return f.set[name]
}

func (f *FlagSet) KVVar(p *map[string]string, name string, value map[string]string, usage string) {
	f.Var(newKVValue(value, p), name, usage)
}

func (f *FlagSet) KVSliceVar(p *[]map[string]string, name string, value []map[string]string, usage string) {
	f.Var(newKVSliceValue(value, p), name, usage)
}

func (f *FlagSet) StringSliceVar(p *[]string, name string, value []string, usage string) {
	f.Var(newStringSliceValue(value, p), name, usage)
}

// ParseFlags parses command line arguments and provides fallback
// values from environment variables and config file values.
// Environment variables are case-insensitive and can have either
// of the provided prefixes.
func (f *FlagSet) ParseFlags(args, environ, prefixes []string, p *properties.Properties) error {
	if err := f.Parse(args); err != nil {
		return err
	}

	if len(prefixes) == 0 {
		prefixes = []string{""}
	}

	// parse environment in case-insensitive way
	env := map[string]string{}
	for _, e := range environ {
		p := strings.SplitN(e, "=", 2)
		env[strings.ToUpper(p[0])] = p[1]
	}

	// determine all values that were set via cmdline
	f.Visit(func(fl *flag.Flag) {
		f.set[fl.Name] = true
	})

	// lookup the rest via environ and properties
	f.VisitAll(func(fl *flag.Flag) {
		// skip if already set
		if f.set[fl.Name] {
			return
		}

		// check environment variables
		for _, pfx := range prefixes {
			name := strings.ToUpper(pfx + strings.Replace(fl.Name, ".", "_", -1))
			if val, ok := env[name]; ok {
				f.set[fl.Name] = true
				f.Set(fl.Name, val)
				return
			}
		}

		// check properties
		if p == nil {
			return
		}
		if val, ok := p.Get(fl.Name); ok {
			f.set[fl.Name] = true
			f.Set(fl.Name, val)
			return
		}
	})
	return nil
}
