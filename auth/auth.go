package auth

import (
	"fmt"
	"net/http"

	"github.com/fabiolb/fabio/config"
)

// AuthScheme defines an authentication scheme.
// See
// - docs/feature/authorization.md
// - docs/ref/proxy.auth.md
type AuthScheme interface {
	// Authorized returns whether request satisfies the authorization scheme.
	Authorized(request *http.Request, response http.ResponseWriter) bool
}

// LoadAuthSchemes takes the 'proxy.auth' configuration option and returns a map
// auth name => auth implementation of a specific type.
func LoadAuthSchemes(cfg map[string]config.AuthScheme) (map[string]AuthScheme, error) {
	auths := map[string]AuthScheme{}
	for _, a := range cfg {
		switch a.Type {
		case "basic":
			b, err := newBasicAuth(a.Basic)
			if err != nil {
				return nil, err
			}
			auths[a.Name] = b
		default:
			return nil, fmt.Errorf("unknown auth type '%s'", a.Type)
		}
	}

	return auths, nil
}
