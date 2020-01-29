package cert

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"path"
	"strings"
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
	Client       *vaultClient
	CertPath     string
	ClientCAPath string
	CAUpgradeCN  string
	Refresh      time.Duration
}

func (s *VaultSource) LoadClientCAs() (*x509.CertPool, error) {
	if s.ClientCAPath == "" {
		return nil, nil
	}
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
	get := func(name, typ string, secret *api.Secret, v2 bool) {
		data := secret.Data
		if v2 {
			x, ok := secret.Data["data"]
			if !ok {
				return
			}
			data, ok = x.(map[string]interface{})
			if !ok {
				return
			}
		}

		v := data[typ]
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

	c, err := s.Client.Get()
	if err != nil {
		return nil, fmt.Errorf("vault: client: %s", err)
	}

	mountPath, v2, err := s.isKVv2(path, c)
	if err != nil {
		return nil, fmt.Errorf("vault: query mount path: %s", err)
	}

	// get the subkeys under 'path'.
	// Each subkey refers to a certificate.
	p := path
	if v2 {
		p = s.addPrefixToVKVPath(p, mountPath, "metadata")
	}

	certs, err := c.Logical().List(p)
	if err != nil {
		return nil, fmt.Errorf("vault: list: %s", err)
	}
	if certs == nil || certs.Data["keys"] == nil {
		return nil, nil
	}

	for _, x := range certs.Data["keys"].([]interface{}) {
		name := x.(string)
		p := path + "/" + name
		if v2 {
			p = s.addPrefixToVKVPath(p, mountPath, "data")
		}
		secret, err := c.Logical().Read(p)
		if err != nil {
			log.Printf("[WARN] cert: Failed to read %s from Vault: %s", p, err)
			continue
		}
		if secret == nil {
			log.Printf("[WARN] cert: Failed to find %s in Vault: %s", p, err)
			continue
		}
		get(name, "cert", secret, v2)
		get(name, "key", secret, v2)
	}

	return pemBlocks, nil
}

func (s *VaultSource) addPrefixToVKVPath(p, mountPath, apiPrefix string) string {
	p = strings.TrimPrefix(p, mountPath)
	return path.Join(mountPath, apiPrefix, p)
}

func (s *VaultSource) isKVv2(path string, client *api.Client) (string, bool, error) {
	mountPath, version, err := s.kvPreflightVersionRequest(client, path)
	if err != nil {
		return "", false, err
	}

	return mountPath, version == 2, nil
}

func (s *VaultSource) kvPreflightVersionRequest(client *api.Client, path string) (string, int, error) {
	r := client.NewRequest("GET", "/v1/sys/internal/ui/mounts/"+path)
	resp, err := client.RawRequest(r)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		// If we get a 404 we are using an older version of vault, default to
		// version 1
		if resp != nil && resp.StatusCode == 404 {
			return "", 1, nil
		}

		return "", 0, err
	}

	secret, err := api.ParseSecret(resp.Body)
	if err != nil {
		return "", 0, err
	}
	var mountPath string
	if mountPathRaw, ok := secret.Data["path"]; ok {
		mountPath = mountPathRaw.(string)
	}
	options := secret.Data["options"]
	if options == nil {
		return mountPath, 1, nil
	}
	versionRaw := options.(map[string]interface{})["version"]
	if versionRaw == nil {
		return mountPath, 1, nil
	}
	version := versionRaw.(string)
	switch version {
	case "", "1":
		return mountPath, 1, nil
	case "2":
		return mountPath, 2, nil
	}

	return mountPath, 1, nil
}
