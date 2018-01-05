package cert

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net/url"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulSource implements a certificate source which loads
// TLS and client authentication certificates from the consul KV store.
// The CertURL/ClientCAURL must point to the base path of the certificates.
// The TLS certificates are updated automatically when the KV store
// changes.
type ConsulSource struct {
	CertURL     string
	ClientCAURL string
	CAUpgradeCN string
}

func parseConsulURL(rawurl string) (config *api.Config, key string, err error) {
	if rawurl == "" || !strings.HasPrefix(rawurl, "http://") && !strings.HasPrefix(rawurl, "https://") {
		return nil, "", errors.New("invalid url")
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, "", err
	}

	config = &api.Config{Address: u.Host, Scheme: u.Scheme}
	if len(u.Query()["token"]) > 0 {
		config.Token = u.Query()["token"][0]
	}

	// path needs to point to kv store and we need
	// to strip the prefix off to get the key
	const prefix = "/v1/kv/"
	key = u.Path
	if !strings.HasPrefix(key, prefix) {
		return nil, "", errors.New("missing prefix: " + prefix)
	}
	key = key[len(prefix):]
	return
}

func (s ConsulSource) LoadClientCAs() (*x509.CertPool, error) {
	if s.ClientCAURL == "" {
		return nil, nil
	}

	config, key, err := parseConsulURL(s.ClientCAURL)
	if err != nil {
		return nil, err
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	load := func(key string) (map[string][]byte, error) {
		pemBlocks, _, err := getCerts(client, key, 0)
		return pemBlocks, err
	}
	return newCertPool(key, s.CAUpgradeCN, load)
}

func (s ConsulSource) Certificates() chan []tls.Certificate {
	if s.CertURL == "" {
		return nil
	}

	config, key, err := parseConsulURL(s.CertURL)
	if err != nil {
		log.Printf("[ERROR] cert: Failed to parse consul url. %s", err)
	}

	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("[ERROR] cert: Failed to create consul client. %s", err)
	}

	pemBlocksCh := make(chan map[string][]byte, 1)
	go watchKV(client, key, pemBlocksCh)

	ch := make(chan []tls.Certificate, 1)
	go func() {
		for pemBlocks := range pemBlocksCh {
			certs, err := loadCertificates(pemBlocks)
			if err != nil {
				log.Printf("[ERROR] cert: Failed to load certificates. %s", err)
				continue
			}
			ch <- certs
		}
	}()
	return ch
}

// watchKV monitors a key in the KV store for changes.
func watchKV(client *api.Client, key string, pemBlocks chan map[string][]byte) {
	var lastIndex uint64
	var lastValue map[string][]byte

	for {
		value, index, err := getCerts(client, key, lastIndex)
		if err != nil {
			log.Printf("[WARN] cert: Error fetching certificates from %s. %v", key, err)
			time.Sleep(time.Second)
			continue
		}

		if !reflect.DeepEqual(value, lastValue) || index != lastIndex {
			log.Printf("[DEBUG] cert: Certificate index changed to #%d", index)
			pemBlocks <- value
			lastValue, lastIndex = value, index
		}
	}
}

func getCerts(client *api.Client, key string, waitIndex uint64) (pemBlocks map[string][]byte, lastIndex uint64, err error) {
	q := &api.QueryOptions{RequireConsistent: true, WaitIndex: waitIndex}
	kvpairs, meta, err := client.KV().List(key, q)
	if err != nil {
		return nil, 0, fmt.Errorf("consul: list: %s", err)
	}
	if len(kvpairs) == 0 {
		return nil, meta.LastIndex, nil
	}
	pemBlocks = map[string][]byte{}
	for _, kvpair := range kvpairs {
		pemBlocks[path.Base(kvpair.Key)] = kvpair.Value
	}
	return pemBlocks, meta.LastIndex, nil
}
