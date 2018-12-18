package prefix

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

const DotDelimiter = "."

const DefaultPrefix = "{{clean .Hostname}}.{{clean .Exec}}"

var (
	prefix string
	once   sync.Once
)

// clean creates safe prefix for graphite reporting by replacing
// some characters with underscores.
// TODO(fs): This may need updating for other metrics backends.
func clean(s string) string {
	if s == "" {
		return "_"
	}
	s = strings.Replace(s, ".", "_", -1)
	s = strings.Replace(s, ":", "_", -1)
	return strings.ToLower(s)
}

func InitPrefix(tmpl string) {
	once.Do(func() {
		// Backward compatibility condition for old metrics.prefix parameter 'default'
		if tmpl == "default" {
			tmpl = DefaultPrefix
		}
		funcMap := template.FuncMap{
			"clean": clean,
		}
		t, err := template.New("prefix").Funcs(funcMap).Parse(tmpl)
		if err != nil {
			panic(err)
		}
		host, err := hostname()
		if err != nil {
			panic(err)
		}
		exe := filepath.Base(os.Args[0])

		b := new(bytes.Buffer)
		data := struct{ Hostname, Exec string }{host, exe}
		if err := t.Execute(b, &data); err != nil {
			panic(err)
		}
		prefix = b.String()
	})
}


func GetPrefix() string {
	return prefix
}

func GetPrefixedName(name string) string {
	if len(prefix) == 0 {
		return name
	}
	return prefix + "." + name
}

// stubbed out for testing
var hostname = os.Hostname
