package noroute

import (
	"sync/atomic"
)

var store atomic.Value // string

func init() {
	store.Store("")
}

// GetHTML returns the HTML for not found routes.
func GetHTML() string {
	return store.Load().(string)
}

// SetHTML sets the HTML for not found routes.
func SetHTML(h string) {
	store.Store(h)
}
