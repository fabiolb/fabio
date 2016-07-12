package cert

import (
	"crypto/tls"
	"crypto/x509"
	"time"
)

// HTTPSource implements a certificate source which loads
// TLS and client authentication certificates from an HTTP/HTTPS server.
// The CertURL/ClientCAURL must point to a text file in the directory
// of the certificates. The text file contains all files that should
// be loaded from this directory - one filename per line.
// The TLS certificates are updated automatically when Refresh
// is not zero. Refresh cannot be less than one second to prevent
// busy loops.
type HTTPSource struct {
	CertURL     string
	ClientCAURL string
	CAUpgradeCN string
	Refresh     time.Duration
}

func (s HTTPSource) LoadClientCAs() (*x509.CertPool, error) {
	return newCertPool(s.ClientCAURL, s.CAUpgradeCN, loadURL)
}

func (s HTTPSource) Certificates() chan []tls.Certificate {
	ch := make(chan []tls.Certificate, 1)
	go watch(ch, s.Refresh, s.CertURL, loadURL)
	return ch
}
