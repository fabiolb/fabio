package cert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
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
	vaultToken string // VAULT_TOKEN env var. Might be wrapped.
	auth       struct {
		// token is the actual Vault token.
		token string

		// expireTime is the time at which the token expires (becomes useless).
		// The zero value indicates that the token is not renewable or never
		// expires.
		expireTime time.Time

		// renewTTL is the desired token lifetime after renewal, in seconds.
		// This value is advisory and the Vault server may ignore or silently
		// change it.
		renewTTL int

		once sync.Once
	}
}

func (s *VaultSource) client() (*api.Client, error) {
	conf := api.DefaultConfig()
	if err := conf.ReadEnvironment(); err != nil {
		return nil, err
	}
	if s.Addr != "" {
		conf.Address = s.Addr
	}
	c, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}

	if err := s.setAuth(c); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *VaultSource) setAuth(c *api.Client) error {
	s.mu.Lock()
	defer func() {
		c.SetToken(s.auth.token)
		s.auth.once.Do(func() { s.checkRenewal(c) })
		s.mu.Unlock()
	}()

	if s.auth.token != "" {
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
			s.auth.token = s.vaultToken
			return nil
		}
		return err
	}
	log.Printf("[INFO] vault: Unwrapped token %s", s.vaultToken)

	s.auth.token = resp.Auth.ClientToken
	return nil
}

// dropNotRenewableWarning controls whether the 'Token is not renewable'
// warning is logged. This is useful for testing where this is the expected
// behavior. On production, this should always be set to false.
var dropNotRenewableWarning bool

// checkRenewal checks if the Vault token can be renewed, and if so when it
// expires and how big the renewal increment should be.
func (s *VaultSource) checkRenewal(c *api.Client) {
	resp, err := c.Auth().Token().LookupSelf()
	if err != nil {
		log.Printf("[WARN] vault: lookup-self failed, token renewal is disabled: %s", err)
		return
	}

	b, _ := json.Marshal(resp.Data)
	var data struct {
		CreationTTL int       `json:"creation_ttl"`
		ExpireTime  time.Time `json:"expire_time"`
		Renewable   bool      `json:"renewable"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		log.Printf("[WARN] vault: lookup-self failed, token renewal is disabled: %s", err)
		return
	}

	switch {
	case data.Renewable:
		s.auth.renewTTL = data.CreationTTL
		s.auth.expireTime = data.ExpireTime
	case data.ExpireTime.IsZero():
		// token doesn't expire
		return
	case dropNotRenewableWarning:
		return
	default:
		ttl := time.Until(data.ExpireTime)
		ttl = ttl / time.Second * time.Second // truncate to seconds
		log.Printf("[WARN] vault: Token is not renewable and will expire %s from now at %s",
			ttl, data.ExpireTime.Format(time.RFC3339))
	}

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

	if err := s.renewToken(c); err != nil {
		log.Printf("[WARN] vault: Failed to renew token: %s", err)
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

func (s *VaultSource) renewToken(c *api.Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ttl := time.Until(s.auth.expireTime)
	switch {
	case s.auth.expireTime.IsZero():
		// Token isn't renewable.
		return nil
	case ttl < 2*s.Refresh:
		// Renew the token if it isn't valid for two more refresh intervals.
		break
	case ttl < 1*time.Minute:
		// Renew the token if it isn't valid for one more minute. This happens
		// if s.Refresh is small, say one second. It is risky to renew the
		// token just one or two seconds before expiration; networks are
		// unreliable, clocks can be skewed, etc.
		break
	default:
		// Token doesn't need to be renewed yet.
		return nil
	}

	resp, err := c.Auth().Token().RenewSelf(s.auth.renewTTL)
	if err != nil {
		return err
	}

	leaseDuration := time.Duration(resp.Auth.LeaseDuration) * time.Second
	s.auth.expireTime = time.Now().Add(leaseDuration)

	return nil
}
