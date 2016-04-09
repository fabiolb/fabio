package cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	vaultapi "github.com/hashicorp/vault/api"
)

type StaticSource struct {
	cert tls.Certificate
}

func (s StaticSource) Certificates() chan []tls.Certificate {
	ch := make(chan []tls.Certificate, 1)
	ch <- []tls.Certificate{s.cert}
	close(ch)
	return ch
}

func (s StaticSource) LoadClientCAs() (*x509.CertPool, error) {
	return nil, nil
}

func TestStaticSource(t *testing.T) {
	certPEM, keyPEM := makeCert("localhost", time.Minute)
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("X509KeyPair: got %s want nil", err)
	}
	testSource(t, StaticSource{cert}, makeCertPool(certPEM), 0)
}

func TestFileSource(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)
	certPEM, keyPEM := makeCert("localhost", time.Minute)
	certFile, keyFile := saveCert(dir, "localhost", certPEM, keyPEM)
	testSource(t, FileSource{CertFile: certFile, KeyFile: keyFile}, makeCertPool(certPEM), 0)
}

func TestPathSource(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)
	certPEM, keyPEM := makeCert("localhost", time.Minute)
	saveCert(dir, "localhost", certPEM, keyPEM)
	testSource(t, PathSource{CertPath: dir}, makeCertPool(certPEM), 0)
}

func TestHTTPSource(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)
	certPEM, keyPEM := makeCert("localhost", time.Minute)
	certFile, keyFile := saveCert(dir, "localhost", certPEM, keyPEM)
	listFile := filepath.Base(certFile) + "\n" + filepath.Base(keyFile) + "\n"
	writeFile(filepath.Join(dir, "list"), []byte(listFile))

	srv := httptest.NewServer(http.FileServer(http.Dir(dir)))
	defer srv.Close()

	testSource(t, HTTPSource{CertURL: srv.URL + "/list"}, makeCertPool(certPEM), 50*time.Millisecond)
}

func TestConsulSource(t *testing.T) {
	const (
		certURL = "http://localhost:8500/v1/kv/fabio/test/consul-server"
		dataDir = "/tmp/fabio-consul-source-test"
	)

	// run a consul server if it isn't already running
	_, err := http.Get("http://localhost:8500/v1/status/leader")
	if err != nil {
		t.Log("Starting consul server")
		consul := exec.Command("consul", "agent", "-server", "-bootstrap", "-data-dir", dataDir)
		if err := consul.Start(); err != nil {
			t.Fatalf("Failed to start consul server. %s", err)
		}
		defer func() {
			consul.Process.Kill()
			os.RemoveAll(dataDir)
		}()

		isUp := func() bool {
			resp, err := http.Get("http://localhost:8500/v1/status/leader")
			return err == nil && resp.StatusCode == 200
		}
		if !waitFor(time.Second, isUp) {
			t.Fatal("Timeout waiting for consul server")
		}
		// give consul time to figure out that it is the only member
		time.Sleep(3 * time.Second)
	} else {
		t.Log("Using existing consul server")
	}

	client, key, err := parseConsulURL(certURL, kvURLPrefix)
	if err != nil {
		t.Fatalf("Failed to create consul client: %s", err)
	}
	defer func() { client.KV().DeleteTree(key, &consulapi.WriteOptions{}) }()

	write := func(name string, value []byte) {
		p := &consulapi.KVPair{Key: key + "/" + name, Value: value}
		_, err := client.KV().Put(p, &consulapi.WriteOptions{})
		if err != nil {
			t.Fatalf("Failed to write %q to consul: %s", p.Key, err)
		}
	}

	certPEM, keyPEM := makeCert("localhost", time.Minute)
	write("localhost-cert.pem", certPEM)
	write("localhost-key.pem", keyPEM)

	testSource(t, ConsulSource{CertURL: certURL}, makeCertPool(certPEM), 50*time.Millisecond)
}

func TestVaultSource(t *testing.T) {
	const (
		addr      = "127.0.0.1:58421"
		rootToken = "token"
		certPath  = "secret/fabio/cert"
	)

	// run a vault server in dev mode
	t.Log("Starting vault server")
	vault := exec.Command("vault", "server", "-dev", "-dev-root-token-id="+rootToken, "-dev-listen-address="+addr)
	if err := vault.Start(); err != nil {
		t.Fatalf("Failed to start vault server. %s", err)
	}
	defer vault.Process.Kill()

	// create a vault client for that server
	c, err := vaultapi.NewClient(&vaultapi.Config{Address: "http://" + addr})
	if err != nil {
		t.Fatalf("NewClient failed: %s", err)
	}
	c.SetToken(rootToken)

	isUp := func() bool {
		ok, err := c.Sys().InitStatus()
		return err == nil && ok
	}
	if !waitFor(time.Second, isUp) {
		t.Fatal("Timeout waiting for vault server")
	}

	// create a renewable token since the vault source
	// will renew the token on every request
	tok, err := c.Auth().Token().Create(&vaultapi.TokenCreateRequest{NoParent: true, TTL: "1h"})
	if err != nil {
		t.Fatalf("Token.Create failed: %s", err)
	}

	// create a cert and store it in vault
	certPEM, keyPEM := makeCert("localhost", time.Minute)
	data := map[string]interface{}{"cert": string(certPEM), "key": string(keyPEM)}
	if _, err := c.Logical().Write(certPath+"/localhost", data); err != nil {
		t.Fatalf("logical.Write failed: %s", err)
	}

	testSource(t, VaultSource{Addr: "http://" + addr, token: tok.Auth.ClientToken, CertPath: certPath}, makeCertPool(certPEM), 50*time.Millisecond)
}

// testSource runs an integration test by making an HTTPS request
// to https://localhost/ expecting that the source provides a valid
// certificate for "localhost". rootCAs is expected to contain a
// valid root certificate or the server certificate itself so that
// the HTTPS client can validate the certificate presented by the
// server.
func testSource(t *testing.T, source Source, rootCAs *x509.CertPool, sleep time.Duration) {
	srvConfig, err := TLSConfig(source)
	if err != nil {
		t.Fatalf("TLSConfig: got %q want nil", err)
	}

	// give the source some time to initialize if necessary
	time.Sleep(sleep)

	// create the https server and start it
	// it will be listening on 127.0.0.1
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	}))
	srv.TLS = srvConfig
	srv.StartTLS()
	defer srv.Close()

	// create an http client that will accept the root CAs
	// otherwise the HTTPS client will not verify the
	// certificate presented by the server.
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			},
		},
	}

	call := func(host string) (statusCode int, body string, err error) {
		// for the certificate validation to work we need to put a hostname
		// which resolves to 127.0.0.1 in the URL. Can't fake the hostname via
		// the Host header.
		resp, err := client.Get(strings.Replace(srv.URL, "127.0.0.1", host, 1))
		if err != nil {
			return 0, "", err
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, "", err
		}

		return resp.StatusCode, string(data), nil
	}

	// disable log output for the next call to prevent
	// confusing log messages since they are expected
	// http: TLS handshake error from 127.0.0.1:55044: remote error: bad certificate
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// make a call for which certificate validation fails.
	// localhost.org is external but resolves to 127.0.0.1
	_, _, err = call("localhost.org")
	if got, want := err, "x509: certificate is valid for localhost, not localhost.org"; got == nil || !strings.Contains(got.Error(), want) {
		t.Fatalf("got %q want %q", got, want)
	}

	// now make the call that should succeed
	statusCode, body, err := call("localhost")
	if err != nil {
		t.Fatalf("got %v want nil", err)
	}
	if got, want := statusCode, 200; got != want {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := body, "OK"; got != want {
		t.Fatalf("got %v want %v", got, want)
	}
}

func tempDir() string {
	dir, err := ioutil.TempDir("", "fabio")
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func writeFile(filename string, data []byte) {
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		log.Fatal(err)
	}
}

func makeCertPool(x ...[]byte) *x509.CertPool {
	p := x509.NewCertPool()
	for _, b := range x {
		p.AppendCertsFromPEM(b)
	}
	return p
}

func saveCert(dir, host string, certPEM, keyPEM []byte) (certFile, keyFile string) {
	certFile, keyFile = filepath.Join(dir, host+"-cert.pem"), filepath.Join(dir, host+"-key.pem")
	writeFile(certFile, certPEM)
	writeFile(keyFile, keyPEM)
	return certFile, keyFile
}

// makeCert creates a self-signed RSA certificate.
// taken from crypto/tls/generate_cert.go
func makeCert(host string, validFor time.Duration) (certPEM, keyPEM []byte) {
	const bits = 1024
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		log.Fatalf("Failed to generate private key: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Fabio Co"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(validFor),
		IsCA:                  true,
		DNSNames:              []string{host},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	var cert, key bytes.Buffer
	pem.Encode(&cert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	pem.Encode(&key, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return cert.Bytes(), key.Bytes()
}

func waitFor(timeout time.Duration, up func() bool) bool {
	until := time.Now().Add(timeout)
	for {
		if time.Now().After(until) {
			return false
		}
		if up() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
}
