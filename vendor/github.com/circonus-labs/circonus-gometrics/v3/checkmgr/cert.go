// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package checkmgr

import (
	"crypto/x509"
	"encoding/json"

	"github.com/pkg/errors"
)

// Default Circonus CA certificate
var circonusCA = []byte(`-----BEGIN CERTIFICATE-----
MIIE6zCCA9OgAwIBAgIJALY0C6uznIh+MA0GCSqGSIb3DQEBCwUAMIGpMQswCQYD
VQQGEwJVUzERMA8GA1UECBMITWFyeWxhbmQxDzANBgNVBAcTBkZ1bHRvbjEXMBUG
A1UEChMOQ2lyY29udXMsIEluYy4xETAPBgNVBAsTCENpcmNvbnVzMSowKAYDVQQD
EyFDaXJjb251cyBDZXJ0aWZpY2F0ZSBBdXRob3JpdHkgRzIxHjAcBgkqhkiG9w0B
CQEWD2NhQGNpcmNvbnVzLm5ldDAeFw0xOTEyMDYyMDAzMzdaFw0zOTEyMDYyMDAz
MzdaMIGpMQswCQYDVQQGEwJVUzERMA8GA1UECBMITWFyeWxhbmQxDzANBgNVBAcT
BkZ1bHRvbjEXMBUGA1UEChMOQ2lyY29udXMsIEluYy4xETAPBgNVBAsTCENpcmNv
bnVzMSowKAYDVQQDEyFDaXJjb251cyBDZXJ0aWZpY2F0ZSBBdXRob3JpdHkgRzIx
HjAcBgkqhkiG9w0BCQEWD2NhQGNpcmNvbnVzLm5ldDCCASIwDQYJKoZIhvcNAQEB
BQADggEPADCCAQoCggEBAK9oN6wBfBgjRYKBbL0Hllcr9TR2e0wIDGhk15Ltym32
zkndEcNKoz61BBJZGalPYDQ8khGQEJAHF6jE/q+qPFHA7vMoIll0frD/C8MM09PK
wvvw+HfnRLjnAWwmefDsE+zhdXlOMnsRPPmMHOCYw0RYe4z8Zna3Jl57zZt8zlKh
FnWRsZg8zc5dFQsAteu2vV+ZSYXUZyj2IgmqaeKgjyUL09ByBKH+weS0ICXiIS51
8lEmofj87ceBMRJHjIwnFr9dRvj3YU/DZVL8NVy91jBHPw9PhLV8XQRh6oQXkrSr
vlcs3NN2FNqWIfZmL6g8/OCCXr3oFgotumGUc7H/cS0CAwEAAaOCARIwggEOMB0G
A1UdDgQWBBRk0xgZQ17grBWWZbRRTzZfqlAd4zCB3gYDVR0jBIHWMIHTgBRk0xgZ
Q17grBWWZbRRTzZfqlAd46GBr6SBrDCBqTELMAkGA1UEBhMCVVMxETAPBgNVBAgT
CE1hcnlsYW5kMQ8wDQYDVQQHEwZGdWx0b24xFzAVBgNVBAoTDkNpcmNvbnVzLCBJ
bmMuMREwDwYDVQQLEwhDaXJjb251czEqMCgGA1UEAxMhQ2lyY29udXMgQ2VydGlm
aWNhdGUgQXV0aG9yaXR5IEcyMR4wHAYJKoZIhvcNAQkBFg9jYUBjaXJjb251cy5u
ZXSCCQC2NAurs5yIfjAMBgNVHRMEBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQCq
9yqOHBWeP65jUnr+pn5nf9+dJhIQ/zgEiIygUwJoSo0+OG1fwfXEeQMQdrYJlTfT
LLgAlK/lJ0fXfS4ruMwyOnH5/2UTrh2eE1u8xToKg7afbaIoO/sg002f3qod1MRx
JYPppNW16wG4kaBKOXJY6LzqXeaStCFotrer5Wt4tl/xOaVav1lmdXC8V3vUtoMJ
FasyBc3tBlgKRJ0f2ijD+P6vEie4w8gJMSurqqKskiY+2zuNzClki0bqCi06m0lt
TESkwBQfV80GJXyz4kTQIZgGnwLcNE9GOlihWX2axTpW7RwpX25lOaMtu+vZtao/
yQRBN07uOh4gEhJIngzr
-----END CERTIFICATE-----`)

// CACert contains cert returned from Circonus API
type CACert struct {
	Contents string `json:"contents"`
}

// loadCACert loads the CA cert for the broker designated by the submission url
func (cm *CheckManager) loadCACert() error {
	if cm.certPool != nil {
		return nil
	}

	if cm.brokerTLS != nil {
		cm.certPool = cm.brokerTLS.RootCAs
		return nil
	}

	cm.certPool = x509.NewCertPool()

	var cert []byte
	var err error

	if cm.enabled {
		// only attempt to retrieve broker CA cert if
		// the check is being managed.
		cert, err = cm.fetchCert()
		if err != nil {
			return err
		}
	}

	if cert == nil {
		cert = circonusCA
	}

	cm.certPool.AppendCertsFromPEM(cert)

	return nil
}

// fetchCert fetches CA certificate using Circonus API
func (cm *CheckManager) fetchCert() ([]byte, error) {
	if !cm.enabled {
		return nil, errors.New("check manager is not enabled")
	}

	cm.Log.Printf("fetching broker cert from api")

	response, err := cm.apih.Get("/pki/ca.crt")
	if err != nil {
		return nil, err
	}

	cadata := new(CACert)
	if err := json.Unmarshal(response, cadata); err != nil {
		return nil, err
	}

	if cadata.Contents == "" {
		return nil, errors.Errorf("error, unable to find ca cert %+v", cadata)
	}

	return []byte(cadata.Contents), nil
}
