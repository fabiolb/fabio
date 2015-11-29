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
	var lastIndex uint64
	var lastValue string

	for {
		value, index, err := getKV(client, path, lastIndex)
		if err != nil {
			log.Printf("[WARN] consul: Error fetching config from %s. %v", path, err)
			time.Sleep(time.Second)
			continue
		}

		if value != lastValue || index != lastIndex {
			log.Printf("[INFO] consul: Manual config changed to #%d", index)
			config <- value
			lastValue, lastIndex = value, index
		}
	}
}

func getKV(client *api.Client, key string, waitIndex uint64) (string, uint64, error) {
	q := &api.QueryOptions{RequireConsistent: true, WaitIndex: waitIndex}
	kvpair, meta, err := client.KV().Get(key, q)
	if err != nil {
		return "", 0, err
	}
	if kvpair == nil {
		return "", meta.LastIndex, nil
	}
	return strings.TrimSpace(string(kvpair.Value)), meta.LastIndex, nil
}

func putKV(client *api.Client, key, value string, index uint64) (bool, error) {
	p := &api.KVPair{Key: key[1:], Value: []byte(value), ModifyIndex: index}
	ok, _, err := client.KV().CAS(p, nil)
	if err != nil {
		return false, err
	}
	return ok, nil
}
