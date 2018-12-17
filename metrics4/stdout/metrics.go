package stdout

import (
	"github.com/fabiolb/fabio/metrics4"
	"github.com/fabiolb/fabio/metrics4/gm"
	rcgm "github.com/rcrowley/go-metrics"
	"log"
	"os"
	"time"
)

func NewProvider(interval time.Duration) metrics4.Provider {
	logger := log.New(os.Stdout, "localhost: ", log.Lmicroseconds)

	r := rcgm.NewRegistry()

	go rcgm.Log(r, interval, logger)

	return gm.NewProvider(r)
}
