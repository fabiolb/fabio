package proxy

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/eBay/fabio/config"
)

func TestAddHeaders(t *testing.T) {
	tests := []struct {
		desc string
		r    *http.Request
		cfg  config.Proxy
		hdrs http.Header
		err  string
	}{
		{"error",
			&http.Request{RemoteAddr: "1.2.3.4"},
			config.Proxy{},
			http.Header{},
			"cannot parse 1.2.3.4",
		},

		{"set remote ip header",
			&http.Request{RemoteAddr: "1.2.3.4:5555"},
			config.Proxy{ClientIPHeader: "Client-IP"},
			http.Header{"Client-Ip": []string{"1.2.3.4"}},
			"",
		},

		{"set remote ip header with local ip (no change expected)",
			&http.Request{RemoteAddr: "1.2.3.4:5555"},
			config.Proxy{LocalIP: "5.6.7.8", ClientIPHeader: "Client-IP"},
			http.Header{"Client-Ip": []string{"1.2.3.4"}},
			"",
		},

		{"set Forwarded",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "example.com:8080"},
			config.Proxy{},
			http.Header{"Forwarded": {"for=1.2.3.4; proto=http; host=example.com:8080"}},
			"",
		},

		{"set Forwarded with localIP",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "example.com:8080"},
			config.Proxy{LocalIP: "5.6.7.8"},
			http.Header{"Forwarded": {"for=1.2.3.4; proto=http; by=5.6.7.8; host=example.com:8080"}},
			"",
		},

		{"set Forwarded with localIP and HTTPS",
			&http.Request{RemoteAddr: "1.2.3.4:5555", TLS: &tls.ConnectionState{}, Host: "example.com:8080"},
			config.Proxy{LocalIP: "5.6.7.8"},
			http.Header{"Forwarded": {"for=1.2.3.4; proto=https; by=5.6.7.8; host=example.com:8080"}},
			"",
		},

		{"extend Forwarded with localIP",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Header: http.Header{"Forwarded": {"for=9.9.9.9; proto=http; by=8.8.8.8"}}, Host: "example.com:8080"},
			config.Proxy{LocalIP: "5.6.7.8"},
			http.Header{"Forwarded": {"for=9.9.9.9; proto=http; by=8.8.8.8; by=5.6.7.8; host=example.com:8080"}},
			"",
		},

		{"set tls header",
			&http.Request{RemoteAddr: "1.2.3.4:5555", TLS: &tls.ConnectionState{}},
			config.Proxy{TLSHeader: "Secure"},
			http.Header{"Secure": {""}},
			"",
		},

		{"set tls header with value",
			&http.Request{RemoteAddr: "1.2.3.4:5555", TLS: &tls.ConnectionState{}},
			config.Proxy{TLSHeader: "Secure", TLSHeaderValue: "true"},
			http.Header{"Secure": {"true"}},
			"",
		},

		{"set X-Forwarded Host and Proto",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "example.com:8080"},
			config.Proxy{},
			http.Header{"X-Forwarded-Host": {"example.com:8080"}, "X-Forwarded-Proto": {"http"}},
			"",
		},
		{"set X-Forwarded Host and Proto with TLS",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "example.com:8080", TLS: &tls.ConnectionState{}},
			config.Proxy{},
			http.Header{"X-Forwarded-Host": {"example.com:8080"}, "X-Forwarded-Proto": {"https"}},
			"",
		},
		{"don't overwrite X-Forwarded-Host header, if present",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "example.com:8080", Header: http.Header{"X-Forwarded-Host": {"original.com"}}},
			config.Proxy{},
			http.Header{"X-Forwarded-Host": {"original.com"}},
			"",
		},
		{"don't overwrite X-Forwarded-Proto header, if present",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Header: http.Header{"X-Forwarded-Proto": {"https"}}},
			config.Proxy{},
			http.Header{"X-Forwarded-Proto": {"https"}},
			"",
		},
		{"set X-Real-Ip header, if not presen",
			&http.Request{RemoteAddr: "1.2.3.4:5555"},
			config.Proxy{},
			http.Header{"X-Real-Ip": {"1.2.3.4"}},
			"",
		},
		{"don't overwrite X-Real-Ip header, if present",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Header: http.Header{"X-Real-Ip": {"6.6.6.6"}}},
			config.Proxy{},
			http.Header{"X-Real-Ip": {"6.6.6.6"}},
			"",
		},
	}

	for i, tt := range tests {
		if tt.r.Header == nil {
			tt.r.Header = http.Header{}
		}

		err := addHeaders(tt.r, tt.cfg)
		if err != nil {
			if got, want := err.Error(), tt.err; got != want {
				t.Errorf("%d: %s\ngot  %q\nwant %q", i, tt.desc, got, want)
			}
			continue
		}
		if tt.err != "" {
			t.Errorf("%d: got nil want %q", i, tt.err)
			continue
		}
		for headerName, _ := range tt.hdrs {
			got := tt.r.Header.Get(headerName)
			want := tt.hdrs.Get(headerName)
			if got != want {
				t.Errorf("%d: %s \nWrong value for Header: %s \ngot  %q \nwant %q", i, tt.desc, headerName, got, want)
			}
		}
	}
}
