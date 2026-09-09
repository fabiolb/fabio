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
	// Authorized returns the authorization decision for request.
	Authorized(request *http.Request, response http.ResponseWriter) AuthDecision
}

// AuthDecision describes whether a request is authorized and whether the
// authentication scheme already completed the client response.
type AuthDecision struct {
	Authorized bool
	Done       bool
}

// LoadAuthSchemes takes the 'proxy.auth' configuration option and returns a map
// auth name => auth implementation of a specific type.
func LoadAuthSchemes(cfg map[string]config.AuthScheme, client *http.Client) (map[string]AuthScheme, error) {
	auths := map[string]AuthScheme{}
	for _, a := range cfg {
		switch a.Type {
		case "basic":
			b, err := newBasicAuth(a.Basic)
			if err != nil {
				return nil, err
			}
			auths[a.Name] = b
		case "external":
			e, err := newExternalAuth(a.External, client)
			if err != nil {
				return nil, err
			}
			auths[a.Name] = e
		default:
			return nil, fmt.Errorf("unknown auth type '%s'", a.Type)
		}
	}

	return auths, nil
}
