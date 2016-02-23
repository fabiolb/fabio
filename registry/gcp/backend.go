// Package gcp implements a backend that be use to
// route traffic to vm-instances that run on the Google Cloud Platform.
package gcp

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/eBay/fabio/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

// FabioKey value must be one of the vm instance tags
const FabioKey = "fabio"

var ErrGoogleCloudPlatformNotConfigured = errors.New("no configuration for Google Cloud Platform backend")

// https://cloud.google.com/compute/docs/metadata?hl=en
type metadataService struct {
	routeInstructions chan string
	computeService    *compute.Service
	config            *config.GoogleCloudPlatform
}

// NewBackend returns a new metadataService that provides a backend implementation
// that periodically queries the metadataservice of the Google Cloud Platform
func NewBackend(cfg *config.GoogleCloudPlatform) (*metadataService, error) {
	if cfg == nil || len(cfg.Project) == 0 {
		return nil, ErrGoogleCloudPlatformNotConfigured
	}
	client, err := google.DefaultClient(oauth2.NoContext, "https://www.googleapis.com/auth/compute")
	if err != nil {
		return nil, err
	}
	computeService, err := compute.New(client)
	if err != nil {
		return nil, err
	}
	service := &metadataService{
		routeInstructions: make(chan string),
		computeService:    computeService,
		config:            cfg,
	}
	go service.poll()
	return service, nil
}

// Register registers fabio as a service in the registry.
func (m *metadataService) Register() error {
	return nil
}

// Deregister removes the service registration for fabio.
func (m *metadataService) Deregister() error {
	return nil
}

// ReadManual returns the current manual overrides and
// their version as seen by the registry.
func (m *metadataService) ReadManual() (value string, version uint64, err error) {
	log.Print("[WARN] manual overrides not supported for GCP")
	return "", 0, nil
}

// WriteManual writes the new value to the registry if the
// version of the stored document still matches version.
func (m *metadataService) WriteManual(value string, version uint64) (ok bool, err error) {
	log.Print("[WARN] manual overrides not supported for GCP")
	return true, nil
}

// WatchServices watches the registry for changes in service
// registration and health and pushes them if there is a difference.
func (m *metadataService) WatchServices() chan string {
	log.Printf("[INFO] gcp: Using routes from metadata values")
	return m.routeInstructions
}

// WatchManual watches the registry for changes in the manual
// overrides and pushes them if there is a difference.
func (m *metadataService) WatchManual() chan string {
	// manual changes can be done via the Google console
	return make(chan string)
}

// https://cloud.google.com/compute/docs/api-rate-limits
func (m *metadataService) poll() {
	for {
		newInstructions := []string{}
		// ask for all running instances
		instances := compute.NewInstancesService(m.computeService)
		// https://cloud.google.com/compute/docs/reference/latest/instances/list
		call := instances.List(m.config.Project, m.config.Zone)
		list, err := call.Do()
		if err != nil {
			log.Printf("[ERROR] get instances failed %v", err)
			goto sleep
		}
		// Todo handle paging
		// for each instance, fetch its metadata
		for _, each := range list.Items {
			getCall := instances.Get(m.config.Project, m.config.Zone, each.Name)
			instance, err := getCall.Do()
			if err != nil {
				log.Printf("[ERROR] get instance failed %v", err)
				goto sleep
			}
			// only running instances can accept traffic
			if "RUNNING" == instance.Status {
				for _, other := range instance.Metadata.Items {
					// for each fabio spec, add the build instruction
					if FabioKey == other.Key && other.Value != nil {
						// invalid instructions are empty
						if entry := buildInstruction(instance, *other.Value); len(entry) > 0 {
							newInstructions = append(newInstructions, entry)
						}
					}
				}
			} else {
				log.Printf("%s instance has status %s", instance.Name, instance.Status)
			}
		}
		// communicate the new instructions (can be empty)
		m.routeInstructions <- strings.Join(newInstructions, "\n")
	sleep:
		log.Printf("%d instructions, waiting %v for the next update", len(newInstructions), m.config.CheckInterval)
		time.Sleep(m.config.CheckInterval)
	}
}
