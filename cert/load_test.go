package cert

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestBase(t *testing.T) {
	tests := []struct {
		in, out, err string
	}{
		{"", "", ""},
		{"http://foo.com/x/y", "http://foo.com/x", ""},
		{"http://foo.com/x/y?p=q", "http://foo.com/x?p=q", ""},
	}

	for i, tt := range tests {
		u, err := base(tt.in)
		if err != nil {
			if got, want := err.Error(), tt.err; got != want {
				t.Errorf("%d: got %v want %v", i, got, want)
				continue
			}
		}
		if tt.err != "" {
			t.Errorf("%d: got nil want %v", i, tt.err)
			continue
		}
		if got, want := u, tt.out; got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}

func TestReplaceSuffix(t *testing.T) {
	if got, want := replaceSuffix("ab", "b", "c"), "ac"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestUpgradeCACertificate(t *testing.T) {
	// generated at
	// https://eu-west-1.console.aws.amazon.com/apigateway/home?region=eu-west-1#/client-certificates
	const awsApiGWCert = `
-----BEGIN CERTIFICATE-----
MIIC6DCCAdCgAwIBAgIIZAgycYqDRqQwDQYJKoZIhvcNAQELBQAwNDELMAkGA1UE
BhMCVVMxEDAOBgNVBAcTB1NlYXR0bGUxEzARBgNVBAMTCkFwaUdhdGV3YXkwHhcN
MTYwNzEyMTkzMTMwWhcNMTcwNzEyMTkzMTMwWjA0MQswCQYDVQQGEwJVUzEQMA4G
A1UEBxMHU2VhdHRsZTETMBEGA1UEAxMKQXBpR2F0ZXdheTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAM/a0LKQd/obIwcKu09EjlHP4b7QqmK/JnJfd1Eq
m6We85FGu+26s7+Bpw1xiyK2jFuzQ4JFyXVkWLJH8e3Mp7P91MvJ1x6UCRk+Fz6Q
Lauw5SBVmDO5CauB4NcICTYEeTT3c0m8t6sDpHf+DHZC87gq9rhBggKXfNO3ntWw
Kq2uGscvnOz2/n2XIucFf2U7GI/cOapGXvIyrB5e/swSCyNkgOJ2HekzWjprxSs5
zu9JSOIzejgm8+/nnPOO9ycVrjN3qazUEXfF1QdvZeNCZ9GL6ZICAYo9xnnNLJnW
6p5d0Fw6U+V/nlNpgCB5djTwXaY51ScoW/i3ukHBZe9QIEcCAwEAATANBgkqhkiG
9w0BAQsFAAOCAQEAzwUJlSv/9XVoeCbot+3mdviZI5B7VnEKGl2Oam1fQzGZkkzB
kqBgtRrHux3BRxPRqS4jM4akdplFhejHExVatOxfS+DEXzFefi+aMb7qApB1YjV/
5FIIQdZaVOlw2KIRXCy04nxrKJmJ1T5RCkYC80dYpNfmDb5REUtp8jU78/Schsx7
0nCsrWkBSO1QtR4NnBlHbEM+imh3aCQz23SUK5Q/NTe4r2pu0zUl5b2YNgefvWle
7fe6T137rmhji9K+tYNznLGk0XmiguQPM2qJLxqeVQsA32wUbbSIFWH+KsXRPfpU
n/iFVG4Y6zyXQY2RzTt+ZB2VPR72X4wqS9fBeQ==
-----END CERTIFICATE-----`

	p, rest := pem.Decode([]byte(awsApiGWCert))
	if len(rest) > 0 {
		t.Fatal("want only one cert")
	}
	cert, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	// check that cert does not have the flags set
	if got, want := cert.BasicConstraintsValid, false; got != want {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := cert.IsCA, false; got != want {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := cert.KeyUsage, x509.KeyUsage(0); got != want {
		t.Fatalf("got %v want %v", got, want)
	}

	// run upgrade with not-matching CN expecting no change
	upgradeCACertificate(cert, "no match")
	if got, want := cert.BasicConstraintsValid, false; got != want {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := cert.IsCA, false; got != want {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := cert.KeyUsage, x509.KeyUsage(0); got != want {
		t.Fatalf("got %v want %v", got, want)
	}

	// run upgrade with matching CN
	upgradeCACertificate(cert, "ApiGateway")
	if got, want := cert.BasicConstraintsValid, true; got != want {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := cert.IsCA, true; got != want {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := cert.KeyUsage, x509.KeyUsageCertSign; got != want {
		t.Fatalf("got %v want %v", got, want)
	}
}
