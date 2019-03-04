package auth

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/tg123/go-htpasswd"
)

// basic is an implementation of AuthScheme
type basic struct {
	realm   string
	secrets *htpasswd.HtpasswdFile
}

func newBasicAuth(cfg config.BasicAuth) (AuthScheme, error) {
	bad := func(err error) {
		log.Println("[WARN] Error processing a line in an htpasswd file:", err)
	}

	stat, err := os.Stat(cfg.File)
	if err != nil {
		return nil, err
	}
	cfg.ModTime = stat.ModTime()

	secrets, err := htpasswd.New(cfg.File, htpasswd.DefaultSystems, bad)
	if err != nil {
		return nil, err
	}

	if cfg.Refresh > 0 {
		go func() {
			ticker := time.NewTicker(cfg.Refresh).C
			for range ticker {
				stat, err := os.Stat(cfg.File)
				if err != nil {
					log.Println("[WARN] Error accessing htpasswd file:", err)
					continue // to prevent nil pointer dereference below
				}

				// refresh the htpasswd file only if its modification time has changed
				// even if the new htpasswd file is older than previously loaded
				if cfg.ModTime != stat.ModTime() {
					if err := secrets.Reload(bad); err == nil {
						log.Println("[INFO] The htpasswd file has been successfully reloaded")
						cfg.ModTime = stat.ModTime()
					} else {
						log.Println("[WARN] Error reloading htpasswd file:", err)
					}
				}
			}
		}()
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
