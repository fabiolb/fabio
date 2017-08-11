package circonus

import (
	"os"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
)

func TestRegistry(t *testing.T) {
	t.Log("Testing registry interface")

	p := &registry{}

	t.Log("\tNames()")
	names := p.Names("")
	if names != nil {
		t.Errorf("Expected nil got '%+v'", names)
	}

	t.Log("\tUnregister()")
	p.Unregister("", "foo")

	t.Log("\tUnregisterAll()")
	p.UnregisterAll("")
}

func TestAll(t *testing.T) {
	start := time.Now()

	if os.Getenv("CIRCONUS_API_TOKEN") == "" {
		t.Skip("skipping test; $CIRCONUS_API_TOKEN not set")
	}

	t.Log("Testing cgm functionality -- this *will* create/use a check")

	cfg := config.Circonus{
		APIKey:   os.Getenv("CIRCONUS_API_TOKEN"),
		APIApp:   os.Getenv("CIRCONUS_API_APP"),
		APIURL:   os.Getenv("CIRCONUS_API_URL"),
		CheckID:  os.Getenv("CIRCONUS_CHECK_ID"),
		BrokerID: os.Getenv("CIRCONUS_BROKER_ID"),
	}

	interval, err := time.ParseDuration("60s")
	if err != nil {
		t.Fatalf("Unable to parse interval %+v", err)
	}

	r, err := NewRegistry("test", cfg, interval)
	if err != nil {
		t.Fatalf("Unable to initialize Circonus +%v", err)
	}

	r.Inc("", "fooCounter", 3)
	r.Time("", "fooTimer", time.Since(start))
	r.metrics.Flush()
}
