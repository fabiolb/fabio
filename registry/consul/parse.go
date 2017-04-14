package consul

import (
	"log"
	"os"
	"strings"
)

// parseURLPrefixTag expects an input in the form of 'tag-host/path[ opts]'
// and returns the lower cased host and the unaltered path if the
// prefix matches the tag.
func parseURLPrefixTag(s, prefix string, env map[string]string) (route, opts string, ok bool) {
	// expand $x or ${x} to env[x] or ""
	expand := func(s string) string {
		return os.Expand(s, func(x string) string {
			if env == nil {
				return ""
			}
			return env[x]
		})
	}

	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, prefix) {
		return "", "", false
	}
	s = strings.TrimSpace(s[len(prefix):])

	p := strings.SplitN(s, " ", 2)
	if len(p) == 2 {
		opts = p[1]
	}
	s = p[0]

	// prefix is ":port"
	if strings.HasPrefix(s, ":") {
		return s, opts, true
	}

	// prefix is "host/path"
	p = strings.SplitN(s, "/", 2)
	if len(p) == 1 {
		log.Printf("[WARN] consul: Invalid %s tag %q - You need to have a trailing slash!", prefix, s)
		return "", "", false
	}
	host, path := p[0], p[1]

	return strings.ToLower(expand(host)) + "/" + expand(path), opts, true
}
