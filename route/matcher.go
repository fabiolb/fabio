package route

import (
	"fmt"
	"log"
	"path"
	"strings"
)

// match contains the matcher function
var match matcher = prefixMatcher

// matcher determines whether a host/path matches a route
type matcher func(uri string, r *Route) bool

// prefixMatcher matches path to the routes' path.
func prefixMatcher(uri string, r *Route) bool {
	return strings.HasPrefix(uri, r.Path)
}

// globMatcher matches path to the routes' path using globbing.
func globMatcher(uri string, r *Route) bool {
	var hasMatch, err = path.Match(r.Path, uri)
	if err != nil {
		log.Printf("[ERROR] Glob matching error %s for path %s route %s", err, uri, r.Path)
		return false
	}
	return hasMatch
}

// SetMatcher sets the matcher function for the proxy.
func SetMatcher(s string) error {
	switch s {
	case "prefix":
		match = prefixMatcher
	case "glob":
		match = globMatcher
	default:
		return fmt.Errorf("route: invalid matcher: %s", s)
	}
	return nil
}
