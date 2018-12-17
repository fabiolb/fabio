package stdout

import (
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/metrics4"
	"github.com/fabiolb/fabio/metrics4/gm"
	rcgm "github.com/rcrowley/go-metrics"
	"log"
	"os"
)

func NewProvider(cfg config.StdOut) (metrics4.Provider, error) {
	logger := log.New(os.Stdout, "localhost: ", log.Lmicroseconds)

	r := rcgm.NewRegistry()

	go rcgm.Log(r, cfg.Interval, logger)

	return gm.NewProvider(r), nil
}

