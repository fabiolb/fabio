package cert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const MaxSize = 1 << 20 // 1MB

func loadURL(listURL string) (pemBlocks map[string][]byte, err error) {
	if listURL == "" {
		return nil, nil
	}

	baseURL, err := base(listURL)
	if err != nil {
		return nil, fmt.Errorf("cert: %s", err)
	}

	fetch := func(url string) (buf []byte, err error) {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}

	// fetch the file with the list of filenames
	list, err := fetch(listURL)
	if err != nil {
		return nil, fmt.Errorf("cert: %s", err)
	}

	// fetch the individual files
	pemBlocks = map[string][]byte{}
	for _, p := range strings.Split(string(list), "\n") {
		if p == "" {
			continue
		}

		path := baseURL + p

		buf, err := fetch(path)
		if err != nil {
			return nil, fmt.Errorf("cert: %s", err)
		}

		pemBlocks[path] = buf
	}

	return pemBlocks, nil
}

func loadPath(root string) (pemBlocks map[string][]byte, err error) {
	if root == "" {
		return nil, nil
	}

	pemBlocks = map[string][]byte{}
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// check if the root directory exists
		if _, ok := err.(*os.PathError); ok && path == root {
			return nil
		}

		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(info.Name()) != ".pem" || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		if info.Size() > MaxSize {
			log.Printf("[WARN] cert: File too large %s", info.Name())
			return nil
		}

		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("cert: %s", err)
		}

		pemBlocks[path] = buf
		return nil
	})

	if err != nil {
		return nil, err
	}

	return pemBlocks, nil
}

func loadCertificates(pemBlocks map[string][]byte) ([]tls.Certificate, error) {
	var n []string
	x := map[string]tls.Certificate{}

	for name := range pemBlocks {
		var certFile, keyFile string
		switch {
		case strings.HasSuffix(name, "-cert.pem"):
			certFile, keyFile = name, replaceSuffix(name, "-cert.pem", "-key.pem")
		case strings.HasSuffix(name, "-key.pem"):
			certFile, keyFile = replaceSuffix(name, "-key.pem", "-cert.pem"), name
		case strings.HasSuffix(name, ".pem"):
			certFile, keyFile = name, name
		default:
			continue
		}

		if _, exists := x[certFile]; exists {
			continue
		}

		cert, key := pemBlocks[certFile], pemBlocks[keyFile]
		if cert == nil || key == nil {
			return nil, fmt.Errorf("cert: cannot load certificate %s", name)
		}

		c, err := tls.X509KeyPair(cert, key)
		if err != nil {
			return nil, fmt.Errorf("cert: invalid certificate %s. %s", name, err)
		}

		x[certFile] = c
		n = append(n, certFile)
	}

	// append certificates in alphabetical order of the
	// cert filenames. This determines which certificate
	// becomes the default certificate (the first one)
	sort.Strings(n)
	var certs []tls.Certificate
	for _, certFile := range n {
		certs = append(certs, x[certFile])
	}

	return certs, nil
}

// base returns the rawurl with the last element of the path
// removed. http://foo.com/x/y becomes http://foo.com/x
func base(rawurl string) (string, error) {
	if rawurl == "" {
		return "", nil
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if u.Path != "/" {
		u.Path = path.Dir(u.Path)
	}
	return u.String(), nil
}

// replaceSuffix replaces oldSuffix with newSuffix in s.
// It is only valid when s has oldSuffix and oldSuffix is not empty.
func replaceSuffix(s string, oldSuffix, newSuffix string) string {
	return s[:len(s)-len(oldSuffix)] + newSuffix
}

// newCertPool creates a new x509.CertPool by loading the
// PEM blocks from loadFn(path) and adding them to a CertPool.
func newCertPool(path string, caUpgradeCN string, loadFn func(path string) (pemBlocks map[string][]byte, err error)) (*x509.CertPool, error) {
	pemBlocks, err := loadFn(path)
	if err != nil {
		return nil, err
	}

	if len(pemBlocks) == 0 {
		return nil, nil
	}

	pool := x509.NewCertPool()
	for _, pemBlock := range pemBlocks {
		for p, rest := pem.Decode(pemBlock); p != nil; p, rest = pem.Decode(rest) {
			cert, err := x509.ParseCertificate(p.Bytes)
			if err != nil {
				return nil, err
			}
			upgradeCACertificate(cert, caUpgradeCN)
			pool.AddCert(cert)
		}
	}

	log.Printf("[INFO] cert: Load client CA certs from %s", path)
	return pool, nil
}

// upgradeCACertificate upgrades a certificate to a self-signing CA certificate if the CN matches.
// Issue #108: Allow generated AWS API Gateway certs to be used for client cert authentication
func upgradeCACertificate(cert *x509.Certificate, caUpgradeCN string) {
	if caUpgradeCN != "" && caUpgradeCN == cert.Issuer.CommonName {
		cert.BasicConstraintsValid = true
		cert.IsCA = true
		cert.KeyUsage = x509.KeyUsageCertSign
		log.Printf("[INFO] cert: Upgrading cert %s to CA cert", cert.Issuer.CommonName)
	}
}
