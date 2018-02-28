package route

import (
	"log"
	"path"
	"strings"

	"github.com/gobwas/glob"
)

// matcher determines whether a host/path matches a route
type matcher func(uri string, r *Route) bool

// Matcher contains the available matcher functions.
// Update config/load.go#load after updating.
var Matcher = map[string]matcher{
	"prefix":      prefixMatcher,
	"glob":        globMatcher,
	"gobwas/glob": gobwasGlobMatcher,
}

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


// gobwasGlobMatcher matches path to the routes' path using gobwas/glob.
func gobwasGlobMatcher(uri string, r *Route) bool {
	var g = glob.MustCompile(r.Path)
	return g.Match(uri)
}
