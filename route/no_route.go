package route

import (
    "log"
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

// SetHTML sets the current noroute html.
func SetHTML(h string) {
    // html := HTML{h}
    store.Store(h)

    if h == "" {
        log.Print("[INFO] Unset noroute HTML")
    } else {
        log.Printf("[INFO] Set noroute HTML (%d bytes)", len(h))
    }
}
