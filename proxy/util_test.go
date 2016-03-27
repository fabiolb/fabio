package proxy

import (
	"crypto/tls"
	"net/http"
	"reflect"
	"testing"

	"github.com/eBay/fabio/config"
	"net"
)

type headerTestset []struct {
	desc string
	r    *http.Request
	cfg  config.Proxy
	hdrs http.Header
}

func Test_addHeaders_ParsingEror(t *testing.T) {
	want := "cannot parse 1.2.3.4"
	got := addHeaders(
		&http.Request{RemoteAddr: "1.2.3.4"},
		config.Proxy{LocalIP: "5.6.7.8"},
	)
	if got.Error() != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func Test_addHeaders_AllTogether(t *testing.T) {
	request := &http.Request{RemoteAddr: "1.2.3.4:5555", Host: "example.com:8080", TLS: &tls.ConnectionState{}, Header: http.Header{}}
	addHeaders(
		request,
		config.Proxy{LocalIP: "5.6.7.8", ClientIPHeader: "Client-IP", TLSHeader: "Secure", TLSHeaderValue: "true"},
	)

	expected := http.Header{
		"Client-Ip":         {"1.2.3.4"},
		"X-Real-Ip":         {"1.2.3.4"},
		"X-Forwarded-Host":  {"example.com:8080"},
		"X-Forwarded-Proto": {"https"},
		"Forwarded":         {"for=1.2.3.4; proto=https; by=5.6.7.8; host=example.com:8080"},
		"Secure":            {"true"},
	}

	if !reflect.DeepEqual(request.Header, expected) {
		t.Errorf("Headers don't match: \ngot  %v\nwant %v", request.Header, expected)
	}
}

func Test_addRemoteIpHeaders(t *testing.T) {
	executeHeaderTests(t, addRemoteIpHeaders,
		headerTestset{
			{"set remote ip header",
				&http.Request{RemoteAddr: "1.2.3.4:5555"},
				config.Proxy{ClientIPHeader: "Client-IP"},
				http.Header{"Client-Ip": {"1.2.3.4"}, "X-Real-Ip": {"1.2.3.4"}},
			},

			{"X-Forwarded-For is skipped as client ip header, because it will be set by golang proxy, later",
				&http.Request{RemoteAddr: "1.2.3.4:5555"},
				config.Proxy{ClientIPHeader: "X-Forwarded-For"},
				http.Header{"X-Real-Ip": {"1.2.3.4"}},
			},

			{"set remote ip header with local ip (no change expected)",
				&http.Request{RemoteAddr: "1.2.3.4:5555"},
				config.Proxy{LocalIP: "5.6.7.8", ClientIPHeader: "Client-IP"},
				http.Header{"Client-Ip": {"1.2.3.4"}, "X-Real-Ip": {"1.2.3.4"}},
			},
		})
}

func Test_addXForwardedHeaders(t *testing.T) {
	executeHeaderTests(t, addXForwardedHeaders,
		headerTestset{
			{"set X-Forwarded Host and Proto",
				&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "example.com:8080"},
				config.Proxy{},
				http.Header{"X-Forwarded-Host": {"example.com:8080"}, "X-Forwarded-Proto": {"http"}},
			},
			{"set X-Forwarded Host and Proto with TLS",
				&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "example.com:8080", TLS: &tls.ConnectionState{}},
				config.Proxy{},
				http.Header{"X-Forwarded-Host": {"example.com:8080"}, "X-Forwarded-Proto": {"https"}},
			},
		})
}

func Test_addForwardedHeader(t *testing.T) {
	executeHeaderTests(t, addForwardedHeader,
		headerTestset{
			{"set Forwarded",
				&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "example.com:8080"},
				config.Proxy{},
				http.Header{"Forwarded": {"for=1.2.3.4; proto=http; host=example.com:8080"}},
			},

			{"set Forwarded with localIP",
				&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "example.com:8080"},
				config.Proxy{LocalIP: "5.6.7.8"},
				http.Header{"Forwarded": {"for=1.2.3.4; proto=http; by=5.6.7.8; host=example.com:8080"}},
			},

			{"set Forwarded with localIP and HTTPS",
				&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "example.com:8080", TLS: &tls.ConnectionState{}},
				config.Proxy{LocalIP: "5.6.7.8"},
				http.Header{"Forwarded": {"for=1.2.3.4; proto=https; by=5.6.7.8; host=example.com:8080"}},
			},

			{"extend Forwarded with localIP",
				&http.Request{RemoteAddr: "1.2.3.4:5555", Host: "example.com:8080", Header: http.Header{"Forwarded": {"for=9.9.9.9; proto=http; by=8.8.8.8"}}},
				config.Proxy{LocalIP: "5.6.7.8"},
				http.Header{"Forwarded": {"for=9.9.9.9; proto=http; by=8.8.8.8; by=5.6.7.8; host=example.com:8080"}},
			},
		})
}

func Test_addTLSHeader(t *testing.T) {
	executeHeaderTests(t, addTLSHeader,
		headerTestset{
			{"set tls header",
				&http.Request{RemoteAddr: "1.2.3.4:5555", TLS: &tls.ConnectionState{}},
				config.Proxy{TLSHeader: "Secure"},
				http.Header{"Secure": {""}},
			},

			{"set tls header with value",
				&http.Request{RemoteAddr: "1.2.3.4:5555", TLS: &tls.ConnectionState{}},
				config.Proxy{TLSHeader: "Secure", TLSHeaderValue: "true"},
				http.Header{"Secure": {"true"}},
			},
		})
}

func executeHeaderTests(t *testing.T, addFunc func(r *http.Request, cfg config.Proxy, remoteIP string), tests headerTestset) {
	for i, tt := range tests {
		remoteIP, _, _ := net.SplitHostPort(tt.r.RemoteAddr)
		if tt.r.Header == nil {
			tt.r.Header = http.Header{}
		}
		addFunc(tt.r, tt.cfg, remoteIP)
		if got, want := tt.r.Header, tt.hdrs; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: %s\ngot  %v\nwant %v", i, tt.desc, got, want)
		}
	}

}
