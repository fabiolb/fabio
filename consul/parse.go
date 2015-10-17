package consul

import (
	"log"
	"strings"
)

// parseURLPrefixTag expects an input in the form of 'tag-host/path'
// and returns the lower cased host plus the path unaltered if the
// prefix matches the tag.
func parseURLPrefixTag(s, prefix string) (host, path string, ok bool) {
	if !strings.HasPrefix(s, prefix) {
		return "", "", false
	}

	// split host/path
	p := strings.SplitN(s[len(prefix):], "/", 2)
	if len(p) != 2 {
		log.Printf("[WARN] Invalid %s tag %q", prefix, s)
		return "", "", false
	}

	host, path = strings.ToLower(strings.TrimSpace(p[0])), "/"+strings.TrimSpace(p[1])
	return host, path, true
}
