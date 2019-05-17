package cert

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/fabiolb/fabio/config"
	"golang.org/x/sync/singleflight"
)

// Source provides the interface for dynamic certificate sources.
type Source interface {
	// Certificates() loads certificates for TLS connections.
	// The first certificate is used as the default certificate
	// if the client does not support SNI or no matching certificate
	// could be found. TLS certificates can be updated at runtime.
	Certificates() chan []tls.Certificate

	// LoadClientCAs() provides certificates for client certificate
	// authentication.
	LoadClientCAs() (*x509.CertPool, error)
}

// Issuer is the interface implemented by sources that can issue certificates
// on-demand.
type Issuer interface {
	// Issue issues a new certificate for the given common name. Issue must
	// return a certificate or an error, never (nil, nil).
	Issue(commonName string) (*tls.Certificate, error)
}

// NewSource generates a cert source from the config options.
func NewSource(cfg config.CertSource) (Source, error) {
	switch cfg.Type {
	case "file":
		return FileSource{
			CertFile:       cfg.CertPath,
			KeyFile:        cfg.KeyPath,
			ClientAuthFile: cfg.ClientCAPath,
			CAUpgradeCN:    cfg.CAUpgradeCN,
		}, nil

	case "path":
		return PathSource{
			CertPath:     cfg.CertPath,
			ClientCAPath: cfg.ClientCAPath,
			CAUpgradeCN:  cfg.CAUpgradeCN,
			Refresh:      cfg.Refresh,
		}, nil

	case "http":
		return HTTPSource{
			CertURL:     cfg.CertPath,
			ClientCAURL: cfg.ClientCAPath,
			CAUpgradeCN: cfg.CAUpgradeCN,
			Refresh:     cfg.Refresh,
		}, nil

	case "consul":
		return ConsulSource{
			CertURL:     cfg.CertPath,
			ClientCAURL: cfg.ClientCAPath,
			CAUpgradeCN: cfg.CAUpgradeCN,
		}, nil

	case "vault":
		return &VaultSource{
			CertPath:     cfg.CertPath,
			ClientCAPath: cfg.ClientCAPath,
			CAUpgradeCN:  cfg.CAUpgradeCN,
			Refresh:      cfg.Refresh,
			Client:       NewVaultClient(cfg.VaultFetchToken),
		}, nil
	case "vault-pki":
		src := NewVaultPKISource()
		src.CertPath = cfg.CertPath
		src.ClientCAPath = cfg.ClientCAPath
		src.CAUpgradeCN = cfg.CAUpgradeCN
		src.Refresh = cfg.Refresh
		src.Client = NewVaultClient(cfg.VaultFetchToken)
		return src, nil

	default:
		return nil, fmt.Errorf("invalid certificate source %q", cfg.Type)
	}
}

// TLSConfig creates a tls.Config which sets the GetCertificate field to a
// certificate store which uses the given source to update the the certificates
// on-demand.
//
// It also sets the ClientCAs field if src.LoadClientCAs returns a non-nil
// value and sets ClientAuth to RequireAndVerifyClientCert.
func TLSConfig(src Source, strictMatch bool, minVersion, maxVersion uint16, cipherSuites []uint16) (*tls.Config, error) {
	clientCAs, err := src.LoadClientCAs()
	if err != nil {
		return nil, err
	}

	sf := &singleflight.Group{}
	store := NewStore()
	x := &tls.Config{
		MinVersion:   minVersion,
		MaxVersion:   maxVersion,
		CipherSuites: cipherSuites,
		NextProtos:   []string{"h2", "http/1.1"},
		GetCertificate: func(clientHello *tls.ClientHelloInfo) (cert *tls.Certificate, err error) {
			cert, err = getCertificate(store.certstore(), clientHello, strictMatch)
			if cert != nil {
				return
			}

			switch err {
			case nil, ErrNoCertsStored:
				// Store doesn't contain a suitable cert. Perhaps the source can issue one?
			default:
				// an unrecoverable error
				return
			}

			ca, ok := src.(Issuer)
			if !ok {
				return
			}

			serverName := clientHello.ServerName
			x, err, _ := sf.Do(serverName, func() (interface{}, error) {
				return ca.Issue(serverName)
			})
			if err != nil {
				return cert, err
			}

			return x.(*tls.Certificate), nil
		},
	}

	if clientCAs != nil {
		x.ClientCAs = clientCAs
		x.ClientAuth = tls.RequireAndVerifyClientCert
	}

	go func() {
		for certs := range src.Certificates() {
			store.SetCertificates(certs)
		}
	}()

	return x, nil
}
