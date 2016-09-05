package cert

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
)

// VaultSource implements a certificate source which loads
// TLS and client authorization certificates from a Vault server.
// The Vault token should be set through the VAULT_TOKEN environment
// variable.
//
// The TLS certificates are updated automatically when Refresh
// is not zero. Refresh cannot be less than one second to prevent
// busy loops.
type VaultSource struct {
	Addr         string
	CertPath     string
	ClientCAPath string
	CAUpgradeCN  string
	Refresh      time.Duration

	mu         sync.Mutex
	token      string // actual token
	vaultToken string // VAULT_TOKEN env var. Might be wrapped.
}

func (s *VaultSource) client() (*api.Client, error) {
	c, err := api.NewClient(&api.Config{Address: s.Addr})
	if err != nil {
		return nil, err
	}
	if err := s.setToken(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *VaultSource) setToken(c *api.Client) error {
	s.mu.Lock()
	defer func() {
		c.SetToken(s.token)
		s.mu.Unlock()
	}()

	if s.token != "" {
		return nil
	}

	if s.vaultToken == "" {
		return errors.New("vault: no token")
	}

	// did we get a wrapped token?
	resp, err := c.Logical().Unwrap(s.vaultToken)
	if err != nil {
		// not a wrapped token?
		if strings.HasPrefix(err.Error(), "no value found at") {
			s.token = s.vaultToken
			return nil
		}
		return err
	}
	log.Printf("[INFO] vault: Unwrapped token %s", s.vaultToken)

	s.token = resp.Auth.ClientToken
	return nil
}

func (s *VaultSource) LoadClientCAs() (*x509.CertPool, error) {
	return newCertPool(s.ClientCAPath, s.CAUpgradeCN, s.load)
}

func (s *VaultSource) Certificates() chan []tls.Certificate {
	ch := make(chan []tls.Certificate, 1)
	go watch(ch, s.Refresh, s.CertPath, s.load)
	return ch
}

func (s *VaultSource) load(path string) (pemBlocks map[string][]byte, err error) {
	pemBlocks = map[string][]byte{}

	// get will read a key=value pair from the secret
	// and store it as <name>-{cert,key}.pem so that
	// they are recognized by the post-processing function
	// which assembles the certificates.
	// The value can be stored either as string or []byte.
	get := func(name, typ string, secret *api.Secret) {
		v := secret.Data[typ]
		if v == nil {
			return
		}

		var b []byte
		switch v.(type) {
		case string:
			b = []byte(v.(string))
		case []byte:
			b = v.([]byte)
		default:
			log.Printf("[WARN] cert: key %s has type %T", name, v)
			return
		}

		pemBlocks[name+"-"+typ+".pem"] = b
	}

	c, err := s.client()
	if err != nil {
		return nil, fmt.Errorf("vault: client: %s", err)
	}

	// renew token
	// TODO(fs): make configurable
	const oneHour = 3600
	_, err = c.Auth().Token().RenewSelf(oneHour)
	if err != nil {
		// TODO(fs): danger of filling up log since default refresh is 1s
		log.Printf("[WARN] vault: Failed to renew token. %s", err)
	}

	// get the subkeys under 'path'.
	// Each subkey refers to a certificate.
	certs, err := c.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("vault: list: %s", err)
	}
	if certs == nil || certs.Data["keys"] == nil {
		return nil, nil
	}

	for _, s := range certs.Data["keys"].([]interface{}) {
		name := s.(string)
		p := path + "/" + name
		secret, err := c.Logical().Read(p)
		if err != nil {
			log.Printf("[WARN] cert: Failed to read %s from Vault: %s", p, err)
			continue
		}
		get(name, "cert", secret)
		get(name, "key", secret)
	}

	return pemBlocks, nil
}
