package registry

type Backend interface {
	// ConfigURL returns the URL to modify the KV store
	// This seems very consul specific and should be fixed
	// with a proper API
	ConfigURL() string

	// WatchServices watches the registry for changes in service
	// registration and health and pushes them if there is a difference.
	WatchServices() chan string

	// WatchManual watches the registry for changes in the manual
	// overrides and pushes them if there is a difference.
	WatchManual() chan string
}
