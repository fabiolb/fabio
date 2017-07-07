package cert

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
)

// vaultClient wraps an *api.Client and takes care of token renewal
// automatically.
type vaultClient struct {
	addr  string // overrides the default config
	token string // overrides the VAULT_TOKEN environment variable

	client *api.Client
	mu     sync.Mutex
}

var DefaultVaultClient = &vaultClient{}

func (c *vaultClient) Get() (*api.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		return c.client, nil
	}

	conf := api.DefaultConfig()
	if err := conf.ReadEnvironment(); err != nil {
		return nil, err
	}

	if c.addr != "" {
		conf.Address = c.addr
	}

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}

	if c.token != "" {
		client.SetToken(c.token)
	}

	token := client.Token()
	if token == "" {
		return nil, errors.New("vault: no token")
	}

	// did we get a wrapped token?
	resp, err := client.Logical().Unwrap(token)
	switch {
	case err == nil:
		log.Printf("[INFO] vault: Unwrapped token %s", token)
		client.SetToken(resp.Auth.ClientToken)
	case strings.HasPrefix(err.Error(), "no value found at"):
		// not a wrapped token
	default:
		return nil, err
	}

	c.client = client
	go c.keepTokenAlive()

	return client, nil
}

// dropNotRenewableWarning controls whether the 'Token is not renewable'
// warning is logged. This is useful for testing where this is the expected
// behavior. On production, this should always be set to false.
var dropNotRenewableWarning bool

func (c *vaultClient) keepTokenAlive() {
	resp, err := c.client.Auth().Token().LookupSelf()
	if err != nil {
		log.Printf("[WARN] vault: lookup-self failed, token renewal is disabled: %s", err)
		return
	}

	b, _ := json.Marshal(resp.Data)
	var data struct {
		TTL         int       `json:"ttl"`
		CreationTTL int       `json:"creation_ttl"`
		Renewable   bool      `json:"renewable"`
		ExpireTime  time.Time `json:"expire_time"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		log.Printf("[WARN] vault: lookup-self failed, token renewal is disabled: %s", err)
		return
	}

	switch {
	case data.Renewable:
		// no-op
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
		return
	}

	ttl := time.Duration(data.TTL) * time.Second
	timer := time.NewTimer(ttl / 2)

	for range timer.C {
		resp, err := c.client.Auth().Token().RenewSelf(data.CreationTTL)
		if err != nil {
			log.Printf("[WARN] vault: Failed to renew token: %s", err)
			timer.Reset(time.Second) // TODO: backoff? abort after N consecutive failures?
			continue
		}

		if !resp.Auth.Renewable || resp.Auth.LeaseDuration == 0 {
			// token isn't renewable anymore, we're done.
			return
		}

		ttl = time.Duration(resp.Auth.LeaseDuration) * time.Second
		timer.Reset(ttl / 2)
	}
}
