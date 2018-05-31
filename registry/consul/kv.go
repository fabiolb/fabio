package consul

import (
	"log"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

// watchKV monitors a key in the KV store for changes.
// The intended use case is to add additional route commands to the routing table.
func watchKV(client *api.Client, path string, config chan string, separator bool) {
	var lastIndex uint64
	var lastValue string

	for {
		value, index, err := listKV(client, path, lastIndex, separator)
		if err != nil {
			log.Printf("[WARN] consul: Error fetching config from %s. %v", path, err)
			time.Sleep(time.Second)
			continue
		}

		if value != lastValue || index != lastIndex {
			log.Printf("[DEBUG] consul: Manual config changed to #%d", index)
			config <- value
			lastValue, lastIndex = value, index
		}
	}
}

func listKeys(client *api.Client, path string, waitIndex uint64) ([]string, uint64, error) {
	q := &api.QueryOptions{RequireConsistent: true, WaitIndex: waitIndex}
	kvpairs, meta, err := client.KV().List(path, q)
	if err != nil {
		return nil, 0, err
	}
	if len(kvpairs) == 0 {
		return nil, meta.LastIndex, nil
	}
	var keys []string
	for _, kvpair := range kvpairs {
		keys = append(keys, kvpair.Key)
	}
	return keys, meta.LastIndex, nil
}

func listKV(client *api.Client, path string, waitIndex uint64, separator bool) (string, uint64, error) {
	q := &api.QueryOptions{RequireConsistent: true, WaitIndex: waitIndex}
	kvpairs, meta, err := client.KV().List(path, q)
	if err != nil {
		return "", 0, err
	}
	if len(kvpairs) == 0 {
		return "", meta.LastIndex, nil
	}
	var s []string
	for _, kvpair := range kvpairs {
		val := strings.TrimSpace(string(kvpair.Value))
		if separator {
			val = "# --- " + kvpair.Key + "\n" + val
		}
		s = append(s, val)
	}
	return strings.Join(s, "\n\n"), meta.LastIndex, nil
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
