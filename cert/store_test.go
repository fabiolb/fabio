package cert

import (
	"crypto/tls"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestGetCertificate(t *testing.T) {
	fooCert := makeCert("foo.com", time.Minute)
	barCert := makeCert("bar.com", time.Minute)
	wildBarCert := makeCert("*.bar.com", time.Minute)

	tests := []struct {
		desc   string
		certs  []tls.Certificate
		hello  *tls.ClientHelloInfo
		strict bool
		cert   *tls.Certificate
		err    error
	}{
		// edge cases
		{
			desc:  "no certs",
			certs: nil,
			hello: &tls.ClientHelloInfo{ServerName: "foo.com"},
			cert:  nil,
			err:   errors.New("cert: no certificates stored"),
		},
		{
			desc:  "server name ends in dot",
			certs: []tls.Certificate{fooCert, barCert},
			hello: &tls.ClientHelloInfo{ServerName: "bar.com."},
			cert:  &barCert,
			err:   nil,
		},

		// happy flows
		{
			desc:  "one cert exact match",
			certs: []tls.Certificate{fooCert},
			hello: &tls.ClientHelloInfo{ServerName: "foo.com"},
			cert:  &fooCert,
			err:   nil,
		},
		{
			desc:  "one cert fallback",
			certs: []tls.Certificate{fooCert},
			hello: &tls.ClientHelloInfo{ServerName: "bar.com"},
			cert:  &fooCert,
			err:   nil,
		},
		{
			desc:   "one cert strict match",
			certs:  []tls.Certificate{fooCert},
			hello:  &tls.ClientHelloInfo{ServerName: "bar.com"},
			cert:   nil,
			strict: true,
			err:    nil,
		},
		{
			desc:  "two certs exact match",
			certs: []tls.Certificate{fooCert, barCert},
			hello: &tls.ClientHelloInfo{ServerName: "bar.com"},
			cert:  &barCert,
			err:   nil,
		},
		{
			desc:  "two certs fallback",
			certs: []tls.Certificate{fooCert, barCert},
			hello: &tls.ClientHelloInfo{ServerName: "whiz.com"},
			cert:  &fooCert,
			err:   nil,
		},
		{
			desc:   "two certs strict match",
			certs:  []tls.Certificate{fooCert, barCert},
			hello:  &tls.ClientHelloInfo{ServerName: "whiz.com"},
			cert:   nil,
			strict: true,
			err:    nil,
		},
		{
			desc:  "wildcard cert",
			certs: []tls.Certificate{fooCert, wildBarCert},
			hello: &tls.ClientHelloInfo{ServerName: "quux.bar.com"},
			cert:  &wildBarCert,
			err:   nil,
		},
		{
			desc:   "wildcard cert strict match",
			certs:  []tls.Certificate{fooCert, wildBarCert},
			hello:  &tls.ClientHelloInfo{ServerName: "quux.bar.com"},
			cert:   &wildBarCert,
			strict: true,
			err:    nil,
		},
	}

	for i, tt := range tests {
		cs := certstore{Certificates: tt.certs}
		cs.BuildNameToCertificate()
		cert, err := getCertificate(cs, tt.hello, tt.strict)
		if got, want := err, tt.err; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: %q: got %v want %v", i, tt.desc, got, want)
			continue
		}
		if got, want := cert, tt.cert; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: %q: got %+v want %+v", i, tt.desc, got, want)
		}
	}
}
