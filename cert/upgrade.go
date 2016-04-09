package cert

import (
	"crypto/x509"
	"encoding/pem"
	"log"
)

func upgrade(pemBlock []byte, commonName string) (pool *x509.CertPool, err error) {
	pool = x509.NewCertPool()
	for p, rest := pem.Decode(pemBlock); p != nil; p, rest = pem.Decode(rest) {
		cert, err := x509.ParseCertificate(p.Bytes)
		if err != nil {
			return nil, err
		}

		// Issue #108: Allow generated AWS API Gateway certs to be used for client cert authentication
		if commonName != "" && commonName == cert.Issuer.CommonName {
			cert.BasicConstraintsValid = true
			cert.IsCA = true
			cert.KeyUsage = x509.KeyUsageCertSign
			log.Print("[INFO] Enabling AWS Api Gateway workaround for certificate %s", cert.Issuer.CommonName)
		}

		pool.AddCert(cert)
	}
	return pool, nil
}
