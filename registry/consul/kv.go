package consul

import (
	"log"
	"strings"
	"time"

	"github.com/eBay/fabio/_third_party/github.com/hashicorp/consul/api"
)

// watchKV monitors a key in the KV store for changes.
// The intended use case is to add addtional route commands to the routing table.
func watchKV(client *api.Client, path string, config chan []string) {
	var lastIndex uint64
	var lastValue string

	for {
		value, index := kvConfig(client, path, lastIndex)
		if value != lastValue || index != lastIndex {
			log.Printf("[INFO] consul: Manual config changed to #%d", index)
			config <- strings.Split(value, "\n")
			lastValue, lastIndex = value, index
		}
	}
}

func kvConfig(client *api.Client, path string, lastIndex uint64) (string, uint64) {
	for {
		q := &api.QueryOptions{RequireConsistent: true, WaitIndex: lastIndex}
		kvpair, meta, err := client.KV().Get(path, q)
		if err != nil {
			log.Printf("[WARN] consul: Error fetching config from %s. %v", path, err)
			time.Sleep(time.Second)
			continue
		}

		if kvpair == nil {
			return "", meta.LastIndex
		}

		return strings.TrimSpace(string(kvpair.Value)), meta.LastIndex
	}
}
