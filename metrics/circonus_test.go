package metrics

import (
	"os"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
)

func TestRegistry(t *testing.T) {
	t.Log("Testing registry interface")

	p := &cgmRegistry{}

	t.Log("\tNames()")
	names := p.Names()
	if names != nil {
		t.Errorf("Expected nil got '%+v'", names)
	}

	t.Log("\tUnregister()")
	p.Unregister("foo")

	t.Log("\tUnregisterAll()")
	p.UnregisterAll()

	t.Log("\tGetTimer()")
	timer := p.GetTimer("foo")
	if timer == nil {
		t.Error("Expected a timer, got nil")
	}
}

func TestTimer(t *testing.T) {
	t.Log("Testing timer interface")

	timer := &cgmTimer{}

	t.Log("\tPercentile()")
	pct := timer.Percentile(99.9)
	if pct != 0 {
		t.Errorf("Expected 0 got '%+v'", pct)
	}

	t.Log("\tRate1()")
	rate := timer.Rate1()
	if rate != 0 {
		t.Errorf("Expected 0 got '%+v'", rate)
	}
}

func TestAll(t *testing.T) {
	start := time.Now()

	if os.Getenv("CIRCONUS_API_TOKEN") == "" && os.Getenv("CIRCONUS_SUBMISSION_URL") == "" {
		t.Skip("skipping test; $CIRCONUS_API_TOKEN or $CIRCONUS_SUBMISSION_URL not set")
	}

	t.Log("Testing cgm functionality -- this *will* create/use a check")

	cfg := config.Circonus{
		SubmissionURL: os.Getenv("CIRCONUS_SUBMISSION_URL"),
		APIKey:        os.Getenv("CIRCONUS_API_TOKEN"),
		APIApp:        os.Getenv("CIRCONUS_API_APP"),
		APIURL:        os.Getenv("CIRCONUS_API_URL"),
		CheckID:       os.Getenv("CIRCONUS_CHECK_ID"),
		BrokerID:      os.Getenv("CIRCONUS_BROKER_ID"),
	}

	interval, err := time.ParseDuration("60s")
	if err != nil {
		t.Fatalf("Unable to parse interval %+v", err)
	}

	circ, err := circonusRegistry("test", cfg, interval)
	if err != nil {
		t.Fatalf("Unable to initialize Circonus +%v", err)
	}

	counter := circ.GetCounter("fooCounter")
	counter.Inc(3)

	timer := circ.GetTimer("fooTimer")
	timer.UpdateSince(start)

	circonus.metrics.Flush()
}
