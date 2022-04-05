package config

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/magiconair/properties"
)

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

type floatSliceValue []float64

func newFloatSliceValue(val []float64, p *[]float64) *floatSliceValue {
	*p = val
	return (*floatSliceValue)(p)
}

func (f *floatSliceValue) String() string {
	strs := make([]string, len(*f))
	for i, v := range *f {
		strs[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return strings.Join(strs, ",")
}

func (f *floatSliceValue) Set(s string) error {
	*f = []float64{}
	for _, x := range strings.Split(s, ",") {
		x = strings.TrimSpace(x)
		if x == "" {
			continue
		}
		v, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return fmt.Errorf("error parsing float slice value %s: %w", x, err)
		}
		*f = append(*f, v)
	}
	return nil
}

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

func (f *FlagSet) StringSliceVar(p *[]string, name string, value []string, usage string) {
	f.Var(newStringSliceValue(value, p), name, usage)
}

func (f *FlagSet) FloatSliceVar(p *[]float64, name string, value []float64, usage string) {
	f.Var(newFloatSliceValue(value, p), name, usage)
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
