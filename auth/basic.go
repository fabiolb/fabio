package auth

import (
	"log"
	"net/http"

	"github.com/fabiolb/fabio/config"
	"github.com/tg123/go-htpasswd"
)

// basic is an implementation of AuthScheme
type basic struct {
	realm   string
	secrets *htpasswd.HtpasswdFile
}

func newBasicAuth(cfg config.BasicAuth) (AuthScheme, error) {
	secrets, err := htpasswd.New(cfg.File, htpasswd.DefaultSystems, func(err error) {
		log.Println("[WARN] Error reading Htpasswd file: ", err)
	})

	if err != nil {
		return nil, err
	}

	return &basic{
		secrets: secrets,
		realm:   cfg.Realm,
	}, nil
}

func (b *basic) Authorized(request *http.Request, response http.ResponseWriter) bool {
	user, password, ok := request.BasicAuth()

	if !ok {
		response.Header().Set("WWW-Authenticate", "Basic realm=\""+b.realm+"\"")
		return false
	}

	return b.secrets.Match(user, password)
}
