package cert

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"strings"
	"sync/atomic"
)

// Store provides a dynamic certificate store which can be updated at
// runtime and is safe for concurrent use.
type Store struct {
	cfg atomic.Value
}

// NewStore creates an empty certificate store.
func NewStore() *Store {
	s := new(Store)
	s.cfg.Store(config{})
	return s
}

// SetCertificates replaces the certificates of the store.
func (s *Store) SetCertificates(certs []tls.Certificate) {
	cfg := config{Certificates: certs}
	cfg.BuildNameToCertificate()
	s.cfg.Store(cfg)
	var names []string
	for name := range cfg.NameToCertificate {
		names = append(names, name)
	}
	log.Printf("[INFO] cert: Store has certificates for [%q]", strings.Join(names, ","))
}

// GetCertificate returns a matching certificate for the given clientHello if possible
// or the first certificate from the store.
func (s *Store) GetCertificate(clientHello *tls.ClientHelloInfo) (cert *tls.Certificate, err error) {
	return getCertificate(s.cfg.Load().(config), clientHello)
}

func getCertificate(cfg config, clientHello *tls.ClientHelloInfo) (cert *tls.Certificate, err error) {
	if len(cfg.Certificates) == 0 {
		return nil, errors.New("cert: no certificates configured")
	}

	if len(cfg.Certificates) == 1 || cfg.NameToCertificate == nil {
		// There's only one choice, so no point doing any work.
		return &cfg.Certificates[0], nil
	}

	name := strings.ToLower(clientHello.ServerName)
	for len(name) > 0 && name[len(name)-1] == '.' {
		name = name[:len(name)-1]
	}

	if cert, ok := cfg.NameToCertificate[name]; ok {
		return cert, nil
	}

	// try replacing labels in the name with wildcards until we get a
	// match.
	labels := strings.Split(name, ".")
	for i := range labels {
		labels[i] = "*"
		candidate := strings.Join(labels, ".")
		if cert, ok := cfg.NameToCertificate[candidate]; ok {
			return cert, nil
		}
	}

	// If nothing matches, return the first certificate.
	return &cfg.Certificates[0], nil
}

type config struct {
	Certificates      []tls.Certificate
	NameToCertificate map[string]*tls.Certificate
}

// BuildNameToCertificate parses Certificates and builds NameToCertificate
// from the CommonName and SubjectAlternateName fields of each of the leaf
// certificates.
func (c *config) BuildNameToCertificate() {
	c.NameToCertificate = make(map[string]*tls.Certificate)
	for i := range c.Certificates {
		cert := &c.Certificates[i]
		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			continue
		}
		if len(x509Cert.Subject.CommonName) > 0 {
			c.NameToCertificate[x509Cert.Subject.CommonName] = cert
		}
		for _, san := range x509Cert.DNSNames {
			c.NameToCertificate[san] = cert
		}
	}
}
