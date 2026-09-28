package auth

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/tg123/go-htpasswd"
)

// basic implements Basic HTTP Authentication.
// It satisfies interface [AuthScheme].
type basic struct {
	secrets *htpasswd.File
	realm   string
	//
	done chan struct{} // USE ONLY FOR TESTING.
}

// newBasicAuth creates a [basic] authentication from cfg.
// It might spawn a forever-running goroutine to periodically refresh the htpassd file.
func newBasicAuth(cfg config.BasicAuth) (*basic, error) {
	bad := func(err error) {
		log.Println("[WARN] Error processing htpasswd file:", err)
	}
	secrets, err := htpasswd.New(cfg.File, htpasswd.DefaultSystems, bad)
	if err != nil {
		return nil, err
	}
	basicAuth := &basic{
		secrets: secrets,
		realm:   cfg.Realm,
		done:    make(chan struct{}),
	}
	if cfg.Refresh == 0 {
		// In this case the htpassd file will not be reloaded, we are done.
		return basicAuth, nil
	}

	// Prepare to reload the contents of the htpassd file each cfg.Refresh seconds.

	stat, err := os.Stat(cfg.File)
	if err != nil {
		return nil, err
	}
	cfg.ModTime = stat.ModTime()

	go func() {
		cleared := false
		ticker := time.NewTicker(cfg.Refresh)
		for {
			select {
			case <-ticker.C:
				stat, err := os.Stat(cfg.File)
				if err != nil {
					log.Println("[WARN] Error accessing htpasswd file:", err)
					if !cleared {
						err = secrets.ReloadFromReader(&bytes.Buffer{}, bad)
						if err != nil {
							log.Println("[WARN] Error clearing the htpasswd credentials:", err)
						} else {
							log.Println("[INFO] The htpasswd credentials have been cleared")
							cleared = true
						}
					}
					continue
				}
				// refresh the htpasswd file only if its modification time has changed
				// even if the new htpasswd file is older than previously loaded
				if cfg.ModTime != stat.ModTime() {
					if err := secrets.Reload(bad); err == nil {
						log.Println("[INFO] The htpasswd file has been successfully reloaded")
						cfg.ModTime = stat.ModTime()
						cleared = false
					} else {
						log.Println("[WARN] Error reloading htpasswd file:", err)
					}
				}
			// A ticker can be stopped but not closed, so we need another channel.
			// USE ONLY FOR TESTING.
			case <-basicAuth.done:
				return
			}
		}
	}()

	return basicAuth, nil
}

// Authorized returns whether request satisfies the authorization scheme.
func (b *basic) Authorized(request *http.Request, response http.ResponseWriter) bool {
	user, password, ok := request.BasicAuth()

	if !ok {
		response.Header().Set("WWW-Authenticate", "Basic realm=\""+b.realm+"\"")
		return false
	}

	return b.secrets.Match(user, password)
}
