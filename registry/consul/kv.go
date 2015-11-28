package consul

import (
	"log"
	"strings"
	"time"

	"github.com/eBay/fabio/_third_party/github.com/hashicorp/consul/api"
)

// watchKV monitors a key in the KV store for changes.
// The intended use case is to add addtional route commands to the routing table.
func watchKV(client *api.Client, path string, config chan string) {
	var index, lastIndex uint64
	var value, lastValue string

	for {
		q := &api.QueryOptions{RequireConsistent: true, WaitIndex: lastIndex}
		kvpair, meta, err := client.KV().Get(path, q)
		if err != nil {
			log.Printf("[WARN] consul: Error fetching config from %s. %v", path, err)
			time.Sleep(time.Second)
			continue
		}

		value, index = "", meta.LastIndex
		if kvpair != nil {
			value, index = strings.TrimSpace(string(kvpair.Value)), meta.LastIndex
		}

		if value != lastValue || index != lastIndex {
			log.Printf("[INFO] consul: Manual config changed to #%d", index)
			config <- value
			lastValue, lastIndex = value, index
		}
	}
}
