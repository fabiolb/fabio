package route

import (
	"net"
	"net/http"
	"net/url"
	"testing"
)

func TestAccessRules_parseAccessRule(t *testing.T) {
	tests := []struct {
		desc      string
		allowDeny string
		fail      bool
	}{
		{
			desc:      "valid ipv4 rule",
			allowDeny: "ip:10.0.0.0/8,ip:192.168.0.0/24,ip:1.2.3.4/32",
		},
		{
			desc:      "valid ipv6 rule",
			allowDeny: "ip:1234:567:beef:cafe::/64,ip:1234:5678:dead:beef::/32",
		},
		{
			desc:      "invalid rule type",
			allowDeny: "xxx:10.0.0.0/8",
			fail:      true,
		},
		{
			desc:      "ip rule with incomplete address",
			allowDeny: "ip:10/8",
			fail:      true,
		},
		{
			desc:      "ip rule with bad cidr mask",
			allowDeny: "ip:10.0.0.0/255",
			fail:      true,
		},
		{
			desc:      "single ipv4 with no mask",
			allowDeny: "ip:1.2.3.4",
			fail:      false,
		},
		{
			desc:      "single ipv6 with no mask",
			allowDeny: "ip:fe80::1",
			fail:      false,
		},
	}

	for i, tt := range tests {
		tt := tt // capture loop var

		t.Run(tt.desc, func(t *testing.T) {
			for _, ad := range []string{"allow", "deny"} {
				tgt := &Target{Opts: map[string]string{ad: tt.allowDeny}}
				err := tgt.parseAccessRule(ad)
				if err != nil && !tt.fail {
					t.Errorf("%d: %s\nfailed: %s", i, tt.desc, err.Error())
					return
				}
			}
		})
	}
}

func TestAccessRules_denyByIP(t *testing.T) {
	tests := []struct {
		desc   string
		target *Target
		remote net.IP
		denied bool
	}{
		{
			desc: "allow rule with included ipv4",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			remote: net.ParseIP("10.10.0.1"),
			denied: false,
		},
		{
			desc: "allow rule with exluded ipv4",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			remote: net.ParseIP("1.2.3.4"),
			denied: true,
		},
		{
			desc: "deny rule with included ipv4",
			target: &Target{
				Opts: map[string]string{"deny": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			remote: net.ParseIP("10.10.0.1"),
			denied: true,
		},
		{
			desc: "deny rule with excluded ipv4",
			target: &Target{
				Opts: map[string]string{"deny": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			remote: net.ParseIP("1.2.3.4"),
			denied: false,
		},
		{
			desc: "allow rule with included ipv6",
			target: &Target{
				Opts: map[string]string{"allow": "ip:1234:dead:beef:cafe::/64"},
			},
			remote: net.ParseIP("1234:dead:beef:cafe::5678"),
			denied: false,
		},
		{
			desc: "allow rule with exluded ipv6",
			target: &Target{
				Opts: map[string]string{"allow": "ip:1234:dead:beef:cafe::/64"},
			},
			remote: net.ParseIP("1234:5678::1"),
			denied: true,
		},
		{
			desc: "deny rule with included ipv6",
			target: &Target{
				Opts: map[string]string{"deny": "ip:1234:dead:beef:cafe::/64"},
			},
			remote: net.ParseIP("1234:dead:beef:cafe::5678"),
			denied: true,
		},
		{
			desc: "deny rule with excluded ipv6",
			target: &Target{
				Opts: map[string]string{"deny": "ip:1234:dead:beef:cafe::/64"},
			},
			remote: net.ParseIP("1234:5678::1"),
			denied: false,
		},
	}

	for i, tt := range tests {
		tt := tt // capture loop var

		t.Run(tt.desc, func(t *testing.T) {
			if err := tt.target.ProcessAccessRules(); err != nil {
				t.Errorf("%d: %s - failed to process access rules: %s",
					i, tt.desc, err.Error())
			}
			tt.target.URL, _ = url.Parse("http://testing.test/")
			if deny := tt.target.denyByIP(tt.remote); deny != tt.denied {
				t.Errorf("%d: %s\ngot denied: %t\nwant denied: %t\n",
					i, tt.desc, deny, tt.denied)
				return
			}
		})
	}
}

func TestAccessRules_AccessDeniedHTTP(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com/", nil)
	tests := []struct {
		desc   string
		target *Target
		xff    string
		remote string
		denied bool
	}{
		{
			desc: "single denied xff and allowed remote addr",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			xff:    "10.11.12.13, 1.1.1.2, 10.11.12.14",
			remote: "10.11.12.1:65500",
			denied: true,
		},
		{
			desc: "allowed xff and denied remote addr",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			xff:    "10.11.12.13, 1.2.3.4",
			remote: "1.1.1.2:65500",
			denied: true,
		},
		{
			desc: "single allowed xff and allowed remote addr",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			xff:    "10.11.12.13, 1.2.3.4",
			remote: "192.168.0.12:65500",
			denied: true,
		},
		{
			desc: "denied xff and denied remote addr",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			xff:    "1.2.3.4, 10.11.12.13, 10.11.12.14",
			remote: "200.17.18.20:65500",
			denied: true,
		},
		{
			desc: "all allowed xff and allowed remote addr",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			xff:    "10.11.12.13, 10.110.120.130",
			remote: "192.168.0.12:65500",
			denied: false,
		},
	}

	for i, tt := range tests {
		tt := tt // capture loop var

		req.Header = http.Header{"X-Forwarded-For": []string{tt.xff}}
		req.RemoteAddr = tt.remote

		t.Run(tt.desc, func(t *testing.T) {
			if err := tt.target.ProcessAccessRules(); err != nil {
				t.Errorf("%d: %s - failed to process access rules: %s",
					i, tt.desc, err.Error())
			}
			tt.target.URL = mustParse("http://testing.test/")
			if deny := tt.target.AccessDeniedHTTP(req); deny != tt.denied {
				t.Errorf("%d: %s\ngot denied: %t\nwant denied: %t\n",
					i, tt.desc, deny, tt.denied)
				return
			}
		})
	}
}
