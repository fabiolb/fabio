package proxy

import (
	"fmt"
	"github.com/eBay/fabio/logger"
	"io"
	"log"
	"os"
)

func newLogger(target string, format string) (*logger.Logger, error) {
	var w io.Writer

	switch target {
	case "stdout":
		log.Printf("[INFO] Output logger to stdout")
		w = os.Stdout
	case "":
		log.Printf("[INFO] Logger disabled")
		return nil, nil
	default:
		return nil, fmt.Errorf("Invalid Logger target %s", target)
	}

	return logger.New(w, format)
}
