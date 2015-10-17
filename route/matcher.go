package route

import "strings"

// match contains the matcher function
var match matcher = prefixMatcher

// matcher determines whether a host/path matches a route
type matcher func(path string, r *route) bool

// prefixMatcher matches path to the routes' path.
func prefixMatcher(path string, r *route) bool {
	return strings.HasPrefix(path, r.path)
}
