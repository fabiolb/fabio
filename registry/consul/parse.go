package consul

import (
	"log"
	"os"
	"strings"
)

// parseURLPrefixTag expects an input in the form of 'tag-host/path'
// and returns the lower cased host plus the path unaltered if the
// prefix matches the tag.
func parseURLPrefixTag(s, prefix string, env map[string]string) (host, path string, ok bool) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, prefix) {
		return "", "", false
	}

	// split host/path
	p := strings.SplitN(s[len(prefix):], "/", 2)
	if len(p) != 2 {
		log.Printf("[WARN] consul: Invalid %s tag %q - You need to have a trailing slash!", prefix, s)
		return "", "", false
	}

	// expand $x or ${x} to env[x] or ""
	expand := func(s string) string {
		return os.Expand(s, func(x string) string {
			if env == nil {
				return ""
			}
			return env[x]
		})
	}

	host = strings.ToLower(expand(strings.TrimSpace(p[0])))
	path = "/" + expand(strings.TrimSpace(p[1]))

	return host, path, true
}
