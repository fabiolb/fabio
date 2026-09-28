package config

import (
	"testing"
)

func TestSanitise(t *testing.T) {
	cfg := &Config{
		Registry: Registry{
			Consul: Consul{
				Addr:  "127.0.0.1:8500",
				Token: "secret-consul-token",
			},
		},
		Metrics: Metrics{
			Circonus: Circonus{
				APIKey: "secret-circonus-key",
				APIApp: "fabio",
			},
		},
		BGP: BGP{
			Peers: []BGPPeer{
				{NeighborAddress: "10.0.0.1", Password: "secret-bgp-password"},
				{NeighborAddress: "10.0.0.2", Password: ""},
			},
		},
		Listen: []Listen{
			{Addr: ":9999", CertSource: CertSource{VaultFetchToken: "secret-vault-token"}},
			{Addr: ":8080", CertSource: CertSource{VaultFetchToken: ""}},
		},
	}

	got := Sanitise(cfg)

	// Sensitive fields must be redacted when non-empty.
	if got.Registry.Consul.Token != redacted {
		t.Errorf("Consul.Token: got %q, want %q", got.Registry.Consul.Token, redacted)
	}
	if got.Metrics.Circonus.APIKey != redacted {
		t.Errorf("Circonus.APIKey: got %q, want %q", got.Metrics.Circonus.APIKey, redacted)
	}
	if got.BGP.Peers[0].Password != redacted {
		t.Errorf("BGPPeer[0].Password: got %q, want %q", got.BGP.Peers[0].Password, redacted)
	}
	if got.Listen[0].CertSource.VaultFetchToken != redacted {
		t.Errorf("Listen[0].VaultFetchToken: got %q, want %q", got.Listen[0].CertSource.VaultFetchToken, redacted)
	}

	// Empty secrets must remain empty (not replaced with the redacted marker).
	if got.BGP.Peers[1].Password != "" {
		t.Errorf("BGPPeer[1].Password: got %q, want empty", got.BGP.Peers[1].Password)
	}
	if got.Listen[1].CertSource.VaultFetchToken != "" {
		t.Errorf("Listen[1].VaultFetchToken: got %q, want empty", got.Listen[1].CertSource.VaultFetchToken)
	}

	// Non-sensitive fields must pass through unchanged.
	if got.Registry.Consul.Addr != "127.0.0.1:8500" {
		t.Errorf("Consul.Addr: got %q, want %q", got.Registry.Consul.Addr, "127.0.0.1:8500")
	}
	if got.Metrics.Circonus.APIApp != "fabio" {
		t.Errorf("Circonus.APIApp: got %q, want %q", got.Metrics.Circonus.APIApp, "fabio")
	}
	if got.BGP.Peers[0].NeighborAddress != "10.0.0.1" {
		t.Errorf("BGPPeer[0].NeighborAddress: got %q, want %q", got.BGP.Peers[0].NeighborAddress, "10.0.0.1")
	}

	// Sanitise must not modify the original config.
	if cfg.Registry.Consul.Token != "secret-consul-token" {
		t.Error("Sanitise modified the original Consul.Token")
	}
	if cfg.BGP.Peers[0].Password != "secret-bgp-password" {
		t.Error("Sanitise modified the original BGPPeer[0].Password")
	}
	if cfg.Listen[0].CertSource.VaultFetchToken != "secret-vault-token" {
		t.Error("Sanitise modified the original Listen[0].VaultFetchToken")
	}
}
