package cert

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"sync"
)

type MuxStore struct {
	sources []Source

	mu *sync.Mutex
	// certs is congruent with sources and stores the latest set of
	// certificates for each source. Used to prioritize certificates in order
	// of the source that provdes them.
	certs [][]tls.Certificate
}

func NewMux(sources []Source) *MuxStore {
	return &MuxStore{
		sources: sources,
		mu:      &sync.Mutex{},
		certs:   make([][]tls.Certificate, len(sources)),
	}
}

var _ Source = (*MuxStore)(nil)
var _ Issuer = (*MuxStore)(nil)

func (m *MuxStore) LoadClientCAs() (*x509.CertPool, error) {
	return m.sources[0].LoadClientCAs() // TODO: change signature to return []*x509.Certificate and merge
}

func (m *MuxStore) Issue(commonName string) (*tls.Certificate, error) {
	for _, s := range m.sources {
		i, ok := s.(Issuer)
		if !ok {
			continue
		}

		cert, err := i.Issue(commonName)
		if err != nil {
			log.Printf("[INFO] cert: Source of type %T cannot issue for %s: %v", s, commonName, err)
			continue
		}

		return cert, nil
	}

	return nil, errors.New("no working cert issuer")
}

func (m *MuxStore) Certificates() chan []tls.Certificate {
	ch := make(chan []tls.Certificate)

	for i, s := range m.sources {
		go func(i int) {
			for certs := range s.Certificates() {
				m.merge(ch, i, certs)
			}
		}(i)
	}

	return ch
}

func (m *MuxStore) merge(ch chan []tls.Certificate, i int, certs []tls.Certificate) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.certs[i] = certs

	var merged []tls.Certificate
	seen := make(map[string]struct{})

	for i, certs := range m.certs {
		for _, cert := range certs {
			if cert.Leaf == nil {
				if len(cert.Certificate) == 0 {
					continue
				}
				var err error
				cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
				if err != nil {
					continue
				}
			}

			key := cert.Leaf.Subject.CommonName
			if _, ok := seen[key]; ok {
				log.Printf("[WARN] cert: Duplicate cert for CN %q in source of type %T", key, m.sources[i])
				continue
			}
			seen[key] = struct{}{}

			merged = append(merged, cert)
		}
	}

	ch <- merged
}
