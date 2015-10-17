package consul

import (
	"log"
	"strings"
	"time"

	"github.com/eBay/fabio/_third_party/github.com/hashicorp/consul/api"
)

// watchManualConfig monitors a key in the KV store for changes and passes
// its content unaltered on. The intended use case is to add addtional
// route commands to the routing table.
func watchManualConfig(client *api.Client, path string, config chan []string) {
	var lastIndex uint64
	var lastValue string

	for {
		value, index := nextValue(client, path, lastIndex)
		if value != lastValue || index != lastIndex {
			log.Printf("[INFO] Manual config changed to #%d", index)
			config <- strings.Split(value, "\n")
			lastValue, lastIndex = value, index
		}
	}
}

func nextValue(client *api.Client, path string, lastIndex uint64) (string, uint64) {
	for {
		q := &api.QueryOptions{RequireConsistent: true, WaitIndex: lastIndex}
		kvpair, meta, err := client.KV().Get(path, q)
		if err != nil {
			log.Printf("[WARN] Error fetching config from %s. %v", path, err)
			time.Sleep(time.Second)
			continue
		}

		if kvpair == nil {
			return "", meta.LastIndex
		}

		return strings.TrimSpace(string(kvpair.Value)), meta.LastIndex
	}
}
