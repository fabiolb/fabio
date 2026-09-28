package config

// redacted is the placeholder used in JSON output for sensitive fields that
// are set but must not be exposed via the /api/config endpoint.
const redacted = "*****"

// mask returns redacted if s is non-empty, otherwise "".
// This lets consumers distinguish "configured but hidden" from "not set".
func mask(s string) string {
	if s != "" {
		return redacted
	}
	return ""
}

// Sanitise returns a deep copy of cfg with all sensitive fields replaced by
// the redacted marker. Use this instead of cfg directly when serialising to
// the /api/config endpoint.
func Sanitise(cfg *Config) Config {
	out := *cfg // shallow copy; slices fixed below

	// Registry.Consul.Token
	out.Registry.Consul = cfg.Registry.Consul
	out.Registry.Consul.Token = mask(cfg.Registry.Consul.Token)

	// Metrics.Circonus.APIKey
	out.Metrics.Circonus = cfg.Metrics.Circonus
	out.Metrics.Circonus.APIKey = mask(cfg.Metrics.Circonus.APIKey)

	// BGP peers passwords
	if len(cfg.BGP.Peers) > 0 {
		peers := make([]BGPPeer, len(cfg.BGP.Peers))
		copy(peers, cfg.BGP.Peers)
		for i := range peers {
			peers[i].Password = mask(cfg.BGP.Peers[i].Password)
		}
		out.BGP.Peers = peers
	}

	// Listen cert sources: VaultFetchToken
	if len(cfg.Listen) > 0 {
		listens := make([]Listen, len(cfg.Listen))
		copy(listens, cfg.Listen)
		for i := range listens {
			listens[i].CertSource.VaultFetchToken = mask(cfg.Listen[i].CertSource.VaultFetchToken)
		}
		out.Listen = listens
	}

	return out
}
