package proxy

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/eBay/fabio/config"
	"github.com/pascaldekloe/goe/verify"
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

		{"http request",
			&http.Request{RemoteAddr: "1.2.3.4:5555"},
			config.Proxy{},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=http"},
				"X-Forwarded-Proto": []string{"http"},
				"X-Forwarded-Port":  []string{"80"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"https request",
			&http.Request{RemoteAddr: "1.2.3.4:5555", TLS: &tls.ConnectionState{}},
			config.Proxy{},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=https"},
				"X-Forwarded-Proto": []string{"https"},
				"X-Forwarded-Port":  []string{"443"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"ws request",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Header: http.Header{"Upgrade": {"websocket"}}},
			config.Proxy{},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=ws"},
				"Upgrade":           []string{"websocket"},
				"X-Forwarded-For":   []string{"1.2.3.4"},
				"X-Forwarded-Proto": []string{"ws"},
				"X-Forwarded-Port":  []string{"80"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"wss request",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Header: http.Header{"Upgrade": {"websocket"}}, TLS: &tls.ConnectionState{}},
			config.Proxy{},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=wss"},
				"Upgrade":           []string{"websocket"},
				"X-Forwarded-For":   []string{"1.2.3.4"},
				"X-Forwarded-Proto": []string{"wss"},
				"X-Forwarded-Port":  []string{"443"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"set client ip header",
			&http.Request{RemoteAddr: "1.2.3.4:5555"},
			config.Proxy{ClientIPHeader: "Client-IP"},
			http.Header{
				"Client-Ip":         []string{"1.2.3.4"},
				"Forwarded":         []string{"for=1.2.3.4; proto=http"},
				"X-Forwarded-Proto": []string{"http"},
				"X-Forwarded-Port":  []string{"80"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"set Forwarded with localIP",
			&http.Request{RemoteAddr: "1.2.3.4:5555"},
			config.Proxy{LocalIP: "5.6.7.8"},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=http; by=5.6.7.8"},
				"X-Forwarded-Proto": []string{"http"},
				"X-Forwarded-Port":  []string{"80"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"set Forwarded with localIP for https",
			&http.Request{RemoteAddr: "1.2.3.4:5555", TLS: &tls.ConnectionState{}},
			config.Proxy{LocalIP: "5.6.7.8"},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=https; by=5.6.7.8"},
				"X-Forwarded-Proto": []string{"https"},
				"X-Forwarded-Port":  []string{"443"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"extend Forwarded with localIP",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Header: http.Header{"Forwarded": {"for=9.9.9.9; proto=http; by=8.8.8.8"}}},
			config.Proxy{LocalIP: "5.6.7.8"},
			http.Header{
				"Forwarded":         []string{"for=9.9.9.9; proto=http; by=8.8.8.8; by=5.6.7.8"},
				"X-Forwarded-Proto": []string{"http"},
				"X-Forwarded-Port":  []string{"80"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"set tls header",
			&http.Request{RemoteAddr: "1.2.3.4:5555", TLS: &tls.ConnectionState{}},
			config.Proxy{TLSHeader: "Secure"},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=https"},
				"Secure":            []string{""},
				"X-Forwarded-Proto": []string{"https"},
				"X-Forwarded-Port":  []string{"443"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"set tls header with value",
			&http.Request{RemoteAddr: "1.2.3.4:5555", TLS: &tls.ConnectionState{}},
			config.Proxy{TLSHeader: "Secure", TLSHeaderValue: "true"},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=https"},
				"Secure":            []string{"true"},
				"X-Forwarded-Proto": []string{"https"},
				"X-Forwarded-Port":  []string{"443"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"overwrite tls header for https, when set",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Header: http.Header{"Secure": []string{"on"}}, TLS: &tls.ConnectionState{}},
			config.Proxy{TLSHeader: "Secure", TLSHeaderValue: "true"},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=https"},
				"Secure":            []string{"true"},
				"X-Forwarded-Proto": []string{"https"},
				"X-Forwarded-Port":  []string{"443"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"drop tls header for http, when set",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Header: http.Header{"Secure": []string{"on"}}},
			config.Proxy{TLSHeader: "Secure", TLSHeaderValue: "true"},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=http"},
				"X-Forwarded-Proto": []string{"http"},
				"X-Forwarded-Port":  []string{"80"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"do not overwrite X-Forwarded-Proto, if present",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Header: http.Header{"X-Forwarded-Proto": {"some value"}}},
			config.Proxy{},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=http"},
				"X-Forwarded-Proto": []string{"some value"},
				"X-Forwarded-Port":  []string{"80"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"set X-Forwarded-Port from Host",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "5.6.7.8:1234"},
			config.Proxy{},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=http"},
				"X-Forwarded-Proto": []string{"http"},
				"X-Forwarded-Port":  []string{"1234"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"set X-Forwarded-Port from Host for https",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "5.6.7.8:1234", TLS: &tls.ConnectionState{}},
			config.Proxy{},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=https"},
				"X-Forwarded-Proto": []string{"https"},
				"X-Forwarded-Port":  []string{"1234"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"do not overwrite X-Forwarded-Port header, if present",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Header: http.Header{"X-Forwarded-Port": {"4444"}}},
			config.Proxy{},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=http"},
				"X-Forwarded-Proto": []string{"http"},
				"X-Forwarded-Port":  []string{"4444"},
				"X-Real-Ip":         []string{"1.2.3.4"},
			},
			"",
		},

		{"do not overwrite X-Real-Ip, if present",
			&http.Request{RemoteAddr: "1.2.3.4:5555", Header: http.Header{"X-Real-Ip": {"6.6.6.6"}}},
			config.Proxy{},
			http.Header{
				"Forwarded":         []string{"for=1.2.3.4; proto=http"},
				"X-Forwarded-Proto": []string{"http"},
				"X-Forwarded-Port":  []string{"80"},
				"X-Real-Ip":         []string{"6.6.6.6"},
			},
			"",
		},
	}

	for i, tt := range tests {
		tt := tt // capture loop var

		t.Run(tt.desc, func(t *testing.T) {
			if tt.r.Header == nil {
				tt.r.Header = http.Header{}
			}

			err := addHeaders(tt.r, tt.cfg)
			if err != nil {
				if got, want := err.Error(), tt.err; got != want {
					t.Fatalf("%d: %s\ngot  %q\nwant %q", i, tt.desc, got, want)
				}
				return
			}

			if tt.err != "" {
				t.Fatalf("%d: got nil want %q", i, tt.err)
				return
			}

			got, want := tt.r.Header, tt.hdrs
			verify.Values(t, "", got, want)
		})
	}
}

func TestLocalPort(t *testing.T) {
	tests := []struct {
		r    *http.Request
		port string
	}{
		{nil, ""},
		{&http.Request{Host: ""}, "80"},
		{&http.Request{Host: ":"}, "80"},
		{&http.Request{Host: "1.2.3.4:5678"}, "5678"},
		{&http.Request{Host: "1.2.3.4"}, "80"},
		{&http.Request{Host: "1.2.3.4", TLS: &tls.ConnectionState{}}, "443"},
		{&http.Request{Host: "1.2.3.4:"}, "80"},
		{&http.Request{Host: "1.2.3.4:", TLS: &tls.ConnectionState{}}, "443"},
	}

	for i, tt := range tests {
		if got, want := localPort(tt.r), tt.port; got != want {
			t.Errorf("%d: got %q want %q", i, got, want)
		}
	}
}
