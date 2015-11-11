package registry

type Backend interface {
	// ConfigURL returns the URL to modify the KV store
	// This seems very consul specific and should be fixed
	// with a proper API
	ConfigURL() string

	// Watch watches the services and manual overrides for changes
	// and pushes them if there is a difference.
	Watch() chan string
}
