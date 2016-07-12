package cert

import (
	"crypto/tls"
	"crypto/x509"
	"path/filepath"
	"time"
)

const (
	DefaultCertPath     = "cert"
	DefaultClientCAPath = "clientca"
)

type PathSource struct {
	Path         string
	CertPath     string
	ClientCAPath string
	CAUpgradeCN  string
	Refresh      time.Duration
}

func (s PathSource) LoadClientCAs() (*x509.CertPool, error) {
	path := makePath(s.Path, s.ClientCAPath, DefaultClientCAPath)
	return newCertPool(path, s.CAUpgradeCN, loadPath)
}

func (s PathSource) Certificates() chan []tls.Certificate {
	path := makePath(s.Path, s.CertPath, DefaultCertPath)
	ch := make(chan []tls.Certificate, 1)
	go watch(ch, s.Refresh, path, loadPath)
	return ch
}

func makePath(parent, child, defaultChild string) string {
	if child == "" {
		return filepath.Join(parent, defaultChild)
	}
	return filepath.Join(parent, child)
}
