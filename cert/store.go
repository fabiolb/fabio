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
	cs atomic.Value
}

// NewStore creates an empty certificate store.
func NewStore() *Store {
	s := new(Store)
	s.cs.Store(certstore{})
	return s
}

// SetCertificates replaces the certificates of the store.
func (s *Store) SetCertificates(certs []tls.Certificate) {
	cs := certstore{Certificates: certs}
	cs.BuildNameToCertificate()
	s.cs.Store(cs)
	var names []string
	for name := range cs.NameToCertificate {
		names = append(names, name)
	}
	log.Printf("[INFO] cert: Store has certificates for [%q]", strings.Join(names, ","))
}

// GetCertificate returns a matching certificate for the given clientHello if possible
// or the first certificate from the store.
func (s *Store) GetCertificate(clientHello *tls.ClientHelloInfo) (cert *tls.Certificate, err error) {
	return getCertificate(s.cs.Load().(certstore), clientHello)
}

func getCertificate(cs certstore, clientHello *tls.ClientHelloInfo) (cert *tls.Certificate, err error) {
	if len(cs.Certificates) == 0 {
		return nil, errors.New("cert: no certificates certstoreured")
	}

	if len(cs.Certificates) == 1 || cs.NameToCertificate == nil {
		// There's only one choice, so no point doing any work.
		return &cs.Certificates[0], nil
	}

	name := strings.ToLower(clientHello.ServerName)
	for len(name) > 0 && name[len(name)-1] == '.' {
		name = name[:len(name)-1]
	}

	if cert, ok := cs.NameToCertificate[name]; ok {
		return cert, nil
	}

	// try replacing labels in the name with wildcards until we get a
	// match.
	labels := strings.Split(name, ".")
	for i := range labels {
		labels[i] = "*"
		candidate := strings.Join(labels, ".")
		if cert, ok := cs.NameToCertificate[candidate]; ok {
			return cert, nil
		}
	}

	// If nothing matches, return the first certificate.
	return &cs.Certificates[0], nil
}

type certstore struct {
	Certificates      []tls.Certificate
	NameToCertificate map[string]*tls.Certificate
}

// BuildNameToCertificate parses Certificates and builds NameToCertificate
// from the CommonName and SubjectAlternateName fields of each of the leaf
// certificates.
func (c *certstore) BuildNameToCertificate() {
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
