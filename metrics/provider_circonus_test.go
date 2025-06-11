package metrics

import (
	"os"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
)

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

	circ, err := NewCirconusProvider("test", cfg, interval)
	if err != nil {
		t.Fatalf("Unable to initialize Circonus +%v", err)
	}

	counter := circ.NewCounter("fooCounter")
	counter.Add(3)

	timer := circ.NewHistogram("fooTimer")
	timer.Observe(time.Since(start).Seconds())

	circonus.metrics.Flush()
}
