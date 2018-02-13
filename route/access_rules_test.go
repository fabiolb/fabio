package route

import (
	"net"
	"net/http"
	"testing"
)

func TestAccessRules_parseAccessRule(t *testing.T) {
	tests := []struct {
		desc      string
		allowDeny string
		fail      bool
	}{
		{
			desc:      "parseAccessRuleGood",
			allowDeny: "ip:10.0.0.0/8,ip:192.168.0.0/24,ip:1.2.3.4/32",
		},
		{
			desc:      "parseAccessRuleBadType",
			allowDeny: "x:10.0.0.0/8",
			fail:      true,
		},
		{
			desc:      "parseAccessRuleIncompleteIP",
			allowDeny: "ip:10/8",
			fail:      true,
		},
		{
			desc:      "parseAccessRuleBadCIDR",
			allowDeny: "ip:10.0.0.0/255",
			fail:      true,
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
			desc: "denyByIPAllowAllowed",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			remote: net.ParseIP("10.10.0.1"),
			denied: false,
		},
		{
			desc: "denyByIPAllowDenied",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			remote: net.ParseIP("1.2.3.4"),
			denied: true,
		},
		{
			desc: "denyByIPDenyDenied",
			target: &Target{
				Opts: map[string]string{"deny": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			remote: net.ParseIP("10.10.0.1"),
			denied: true,
		},
		{
			desc: "denyByIPDenyAllow",
			target: &Target{
				Opts: map[string]string{"deny": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			remote: net.ParseIP("1.2.3.4"),
			denied: false,
		},
	}

	for i, tt := range tests {
		tt := tt // capture loop var

		t.Run(tt.desc, func(t *testing.T) {
			if err := tt.target.processAccessRules(); err != nil {
				t.Errorf("%d: %s - failed to process access rules: %s",
					i, tt.desc, err.Error())
			}
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
			desc: "AccessDeniedHTTPwithDeniedXFFandAllowedRemote",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			xff:    "1.1.1.2, 10.11.12.13, 10.11.12.14",
			remote: "10.11.12.1:65500",
			denied: true,
		},
		{
			desc: "AccessDeniedHTTPwithAllowedXFFandDeniedRemote",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			xff:    "10.11.12.13, 1.2.3.4",
			remote: "1.1.1.2:65500",
			denied: true,
		},
		{
			desc: "AccessDeniedHTTPwitAllowedXFFandAllowedRemote",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			xff:    "10.11.12.13, 1.2.3.4",
			remote: "192.168.0.12:65500",
			denied: false,
		},
		{
			desc: "AccessDeniedHTTPwithDeniedXFFandDeniedRemote",
			target: &Target{
				Opts: map[string]string{"allow": "ip:10.0.0.0/8,ip:192.168.0.0/24"},
			},
			xff:    "1.2.3.4, 10.11.12.13, 10.11.12.14",
			remote: "200.17.18.20:65500",
			denied: true,
		},
	}

	for i, tt := range tests {
		tt := tt // capture loop var

		req.Header = http.Header{"X-Forwarded-For": []string{tt.xff}}
		req.RemoteAddr = tt.remote

		t.Run(tt.desc, func(t *testing.T) {
			if err := tt.target.processAccessRules(); err != nil {
				t.Errorf("%d: %s - failed to process access rules: %s",
					i, tt.desc, err.Error())
			}
			if deny := tt.target.AccessDeniedHTTP(req); deny != tt.denied {
				t.Errorf("%d: %s\ngot denied: %t\nwant denied: %t\n",
					i, tt.desc, deny, tt.denied)
				return
			}
		})
	}
}
