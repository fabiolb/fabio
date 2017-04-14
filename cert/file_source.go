package cert

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/fabiolb/fabio/exit"
)

// FileSource implements a certificate source for one
// TLS and one client authentication certificate.
// The certificates are loaded during startup and are cached
// in memory until the program exits.
// It exists to support the legacy configuration only. The
// PathSource should be used instead.
type FileSource struct {
	CertFile       string
	KeyFile        string
	ClientAuthFile string
	CAUpgradeCN    string
}

func (s FileSource) LoadClientCAs() (*x509.CertPool, error) {
	return newCertPool(s.ClientAuthFile, s.CAUpgradeCN, func(path string) (map[string][]byte, error) {
		if s.ClientAuthFile == "" {
			return nil, nil
		}
		pemBlock, err := ioutil.ReadFile(path)
		return map[string][]byte{path: pemBlock}, err
	})
}

func (s FileSource) Certificates() chan []tls.Certificate {
	ch := make(chan []tls.Certificate, 1)
	ch <- []tls.Certificate{loadX509KeyPair(s.CertFile, s.KeyFile)}
	close(ch)
	return ch
}

func loadX509KeyPair(certFile, keyFile string) tls.Certificate {
	if certFile == "" {
		exit.Fatalf("[FATAL] cert: CertFile is required")
	}

	if keyFile == "" {
		keyFile = certFile
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		exit.Fatalf("[FATAL] cert: Error loading certificate. %s", err)
	}
	return cert
}
