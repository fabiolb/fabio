package registry

type Backend interface {
	// Register registers fabio as a service in the registry.
	Register(services []string) error

	// Deregister removes all service registrations for fabio.
	DeregisterAll() error

	// Deregister removes the given service registration for fabio.
	Deregister(service string) error

	// ManualPaths returns the list of paths for which there
	// are overrides.
	ManualPaths() ([]string, error)

	// ReadManual returns the current manual overrides and
	// their version as seen by the registry.
	ReadManual(path string) (value string, version uint64, err error)

	// WriteManual writes the new value to the registry if the
	// version of the stored document still matchhes version.
	WriteManual(path string, value string, version uint64) (ok bool, err error)

	// WatchServices watches the registry for changes in service
	// registration and health and pushes them if there is a difference.
	WatchServices() chan string

	// WatchManual watches the registry for changes in the manual
	// overrides and pushes them if there is a difference.
	WatchManual() chan string

	// WatchNoRouteHTML watches the registry for changes in the html returned
	// when a requested route is not found
	WatchNoRouteHTML() chan string
}

var Default Backend
