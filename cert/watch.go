package cert

import (
	"crypto/tls"
	"github.com/eBay/fabio/mdllog"
	"reflect"
	"time"
)

// watch monitors the result of the loadFn function for changes.
func watch(ch chan []tls.Certificate, refresh time.Duration, path string, loadFn func(path string) (map[string][]byte, error)) {
	once := refresh <= 0

	// do not refresh more often than once a second to prevent busy loops
	if refresh < time.Second {
		refresh = time.Second
	}

	var last map[string][]byte
	for {
		next, err := loadFn(path)
		if err != nil {
			mdllog.Error.Printf("[ERROR] cert: Cannot load certificates from %s. %s", path, err)
			time.Sleep(refresh)
			continue
		}

		if reflect.DeepEqual(next, last) {
			time.Sleep(refresh)
			continue
		}

		certs, err := loadCertificates(next)
		if err != nil {
			mdllog.Error.Printf("[ERROR] cert: Cannot make certificates: %s", err)
			continue
		}

		ch <- certs
		last = next

		if once {
			return
		}
	}
}
