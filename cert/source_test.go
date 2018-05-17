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
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/http2"

	"github.com/fabiolb/fabio/config"
	consulapi "github.com/hashicorp/consul/api"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pascaldekloe/goe/verify"
)

func TestTLSConfig(t *testing.T) {
	certPEM, keyPEM := makePEM("localhost", time.Minute)
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("X509KeyPair: got %s want nil", err)
	}
	pool := makeCertPool(certPEM)
	src := &StaticSource{cert, pool}
	tlsmin := uint16(0x1000)
	tlsmax := uint16(0x2000)
	tlsciphers := []uint16{0x1234, 0x5678}
	nextprotos := []string{"h2", "http/1.1"}

	cfg, err := TLSConfig(src, false, tlsmin, tlsmax, tlsciphers)
	if err != nil {
		t.Fatalf("got error %v want nil", err)
	}
	if got, want := cfg.MinVersion, tlsmin; got != want {
		t.Fatalf("got tls min version %04x want %04x", got, want)
	}
	if got, want := cfg.MaxVersion, tlsmax; got != want {
		t.Fatalf("got tls max version %04x want %04x", got, want)
	}
	if got, want := cfg.CipherSuites, tlsciphers; !reflect.DeepEqual(got, want) {
		t.Fatalf("got tls ciphers %v want %v", got, want)
	}
	if got, want := cfg.NextProtos, nextprotos; !reflect.DeepEqual(got, want) {
		t.Fatalf("got next protos %v want %v", got, want)
	}
	if got, want := cfg.ClientCAs, pool; got != want {
		t.Fatalf("got client CAs %v want %v", got, want)
	}
	if got, want := cfg.ClientAuth, tls.RequireAndVerifyClientCert; got != want {
		t.Fatalf("got client auth type %v want %v", got, want)
	}
	if cfg.GetCertificate == nil {
		t.Fatalf("got GetCertificate() nil want not nil")
	}
}

func TestNewSource(t *testing.T) {
	certsource := func(typ string) config.CertSource {
		return config.CertSource{
			Type:         typ,
			Name:         "name",
			CertPath:     "cert",
			KeyPath:      "key",
			ClientCAPath: "clientca",
			CAUpgradeCN:  "upgcn",
			Refresh:      3 * time.Second,
			Header:       http.Header{"A": []string{"b"}},
		}
	}
	tests := []struct {
		desc string
		cfg  config.CertSource
		src  Source
		err  string
	}{
		{
			desc: "invalid",
			cfg: config.CertSource{
				Type: "invalid",
			},
			src: nil,
			err: `invalid certificate source "invalid"`,
		},
		{
			desc: "file",
			cfg:  certsource("file"),
			src: FileSource{
				CertFile:       "cert",
				KeyFile:        "key",
				ClientAuthFile: "clientca",
				CAUpgradeCN:    "upgcn",
			},
		},
		{
			desc: "path",
			cfg:  certsource("path"),
			src: PathSource{
				CertPath:     "cert",
				ClientCAPath: "clientca",
				CAUpgradeCN:  "upgcn",
				Refresh:      3 * time.Second,
			},
		},
		{
			desc: "http",
			cfg:  certsource("http"),
			src: HTTPSource{
				CertURL:     "cert",
				ClientCAURL: "clientca",
				CAUpgradeCN: "upgcn",
				Refresh:     3 * time.Second,
			},
		},
		{
			desc: "consul",
			cfg:  certsource("consul"),
			src: ConsulSource{
				CertURL:     "cert",
				ClientCAURL: "clientca",
				CAUpgradeCN: "upgcn",
			},
		},
		{
			desc: "vault",
			cfg:  certsource("vault"),
			src: &VaultSource{
				Client:       DefaultVaultClient,
				CertPath:     "cert",
				ClientCAPath: "clientca",
				CAUpgradeCN:  "upgcn",
				Refresh:      3 * time.Second,
			},
		},
	}

	for i, tt := range tests {
		tt := tt // capture loop var
		t.Run(tt.desc, func(t *testing.T) {
			var errmsg string
			src, err := NewSource(tt.cfg)
			if err != nil {
				errmsg = err.Error()
			}
			if got, want := errmsg, tt.err; got != want {
				t.Fatalf("%d: got %q want %q", i, got, want)
			}
			got, want := src, tt.src
			verify.Values(t, "src", got, want)
		})
	}
}

type StaticSource struct {
	cert tls.Certificate
	pool *x509.CertPool
}

func (s StaticSource) Certificates() chan []tls.Certificate {
	ch := make(chan []tls.Certificate, 1)
	ch <- []tls.Certificate{s.cert}
	close(ch)
	return ch
}

func (s StaticSource) LoadClientCAs() (*x509.CertPool, error) {
	return s.pool, nil
}

func TestStaticSource(t *testing.T) {
	certPEM, keyPEM := makePEM("localhost", time.Minute)
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("X509KeyPair: got %s want nil", err)
	}
	testSource(t, StaticSource{cert, nil}, makeCertPool(certPEM), 0)
}

func TestFileSource(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)
	certPEM, keyPEM := makePEM("localhost", time.Minute)
	certFile, keyFile := saveCert(dir, "localhost", certPEM, keyPEM)
	testSource(t, FileSource{CertFile: certFile, KeyFile: keyFile}, makeCertPool(certPEM), 0)
}

func TestPathSource(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)
	certPEM, keyPEM := makePEM("localhost", time.Minute)
	saveCert(dir, "localhost", certPEM, keyPEM)
	testSource(t, PathSource{CertPath: dir}, makeCertPool(certPEM), 10*time.Millisecond)
}

func TestHTTPSource(t *testing.T) {
	dir := tempDir()
	defer os.RemoveAll(dir)
	certPEM, keyPEM := makePEM("localhost", time.Minute)
	certFile, keyFile := saveCert(dir, "localhost", certPEM, keyPEM)
	listFile := filepath.Base(certFile) + "\n" + filepath.Base(keyFile) + "\n"
	writeFile(filepath.Join(dir, "list"), []byte(listFile))

	srv := httptest.NewServer(http.FileServer(http.Dir(dir)))
	defer srv.Close()

	testSource(t, HTTPSource{CertURL: srv.URL + "/list"}, makeCertPool(certPEM), 500*time.Millisecond)
}

func TestConsulSource(t *testing.T) {
	const certURL = "http://127.0.0.1:8500/v1/kv/fabio/test/consul-server"

	// run a consul server if it isn't already running
	_, err := http.Get("http://127.0.0.1:8500/v1/status/leader")
	if err != nil {
		consul := os.Getenv("CONSUL_EXE")
		if consul == "" {
			consul = "consul"
		}

		version, err := exec.Command(consul, "--version").Output()
		if err != nil {
			t.Fatalf("Failed to run %s --version", consul)
		}
		cr := bytes.IndexRune(version, '\n')
		t.Logf("Starting %s: %s", consul, string(version[:cr]))

		start := time.Now()
		cmd := exec.Command(consul, "agent", "-bind", "127.0.0.1", "-server", "-dev")
		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start consul server. %s", err)
		}
		defer cmd.Process.Kill()

		isUp := func() bool {
			resp, err := http.Get("http://127.0.0.1:8500/v1/status/leader")
			// /v1/status/leader returns '\n""' while consul is in leader election mode
			// and '"127.0.0.1:8300"' when not. So we punt by checking the
			// body length instead of the actual body content :)
			if err != nil {
				return false
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				return false
			}

			n, err := io.Copy(ioutil.Discard, resp.Body)
			return err == nil && n > 10
		}

		// We need give consul ~8-10 seconds to become ready until I've
		// figured out whether we can speed this up. Make sure that this is
		// less than the global test timeout in Makefile.
		if !waitFor(12*time.Second, isUp) {
			t.Fatalf("Timeout waiting for consul server after %2.1f seconds", time.Since(start).Seconds())
		}
		t.Logf("Consul is ready after %2.1f seconds", time.Since(start).Seconds())
	} else {
		t.Log("Using existing consul server")
	}

	config, key, err := parseConsulURL(certURL)
	if err != nil {
		t.Fatalf("Failed to parse consul url: %s", err)
	}

	client, err := consulapi.NewClient(config)
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

	certPEM, keyPEM := makePEM("localhost", time.Minute)
	write("localhost-cert.pem", certPEM)
	write("localhost-key.pem", keyPEM)

	testSource(t, ConsulSource{CertURL: certURL}, makeCertPool(certPEM), 500*time.Millisecond)
}

// vaultServer starts a vault server in dev mode and waits until is ready.
func vaultServer(t *testing.T, addr, rootToken string) (*exec.Cmd, *vaultapi.Client) {
	vault := os.Getenv("VAULT_EXE")
	if vault == "" {
		vault = "vault"
	}

	version, err := exec.Command(vault, "--version").Output()
	if err != nil {
		t.Fatalf("Failed to run %s --version", vault)
	}
	t.Logf("Starting %s: %q", vault, string(version))

	cmd := exec.Command(vault, "server", "-dev", "-dev-root-token-id="+rootToken, "-dev-listen-address="+addr)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start vault server. %s", err)
	}

	c, err := vaultapi.NewClient(&vaultapi.Config{Address: "http://" + addr})
	if err != nil {
		cmd.Process.Kill()
		t.Fatalf("NewClient failed: %s", err)
	}
	c.SetToken(rootToken)

	isUp := func() bool {
		ok, err := c.Sys().InitStatus()
		return err == nil && ok
	}
	if !waitFor(time.Second, isUp) {
		cmd.Process.Kill()
		t.Fatal("Timeout waiting for vault server")
	}

	policy := `
	# Vault < 0.7
	path "secret/fabio/cert" {
	  capabilities = ["list"]
	}
	# Vault >= 0.7. Note the trailing slash.
	path "secret/fabio/cert/" {
	  capabilities = ["list"]
	}

	path "secret/fabio/cert/*" {
	  capabilities = ["read"]
	}

	# Vault >= 0.10. (KV Version 2)
	path "secret/metadata/fabio/cert/" {
	  capabilities = ["list"]
	}

	path "secret/data/fabio/cert/*" {
	  capabilities = ["read"]
	}

	path "test-pki/issue/fabio" {
	  capabilities = ["update"]
	}
	`

	if err := c.Sys().PutPolicy("fabio", policy); err != nil {
		cmd.Process.Kill()
		t.Fatalf("Could not create policy: %s", err)
	}

	return cmd, c
}

func makeToken(t *testing.T, c *vaultapi.Client, wrapTTL string, req *vaultapi.TokenCreateRequest) string {
	c.SetWrappingLookupFunc(func(string, string) string { return wrapTTL })

	resp, err := c.Auth().Token().Create(req)
	if err != nil {
		t.Fatalf("Could not create a token: %s", err)
	}

	if wrapTTL != "" {
		if resp.WrapInfo == nil || resp.WrapInfo.Token == "" {
			t.Fatalf("Could not create a wrapped token")
		}
		return resp.WrapInfo.Token
	}

	if resp.WrapInfo != nil && resp.WrapInfo.Token != "" {
		t.Fatalf("Got a wrapped token but was not expecting one")
	}

	return resp.Auth.ClientToken
}

var vaultTestCases = []struct {
	desc     string
	wrapTTL  string
	req      *vaultapi.TokenCreateRequest
	dropWarn bool
}{
	{
		desc: "renewable token",
		req:  &vaultapi.TokenCreateRequest{Lease: "1m", TTL: "1m", Policies: []string{"fabio"}},
	},
	{
		desc:     "non-renewable token",
		req:      &vaultapi.TokenCreateRequest{Lease: "1m", TTL: "1m", Renewable: new(bool), Policies: []string{"fabio"}},
		dropWarn: true,
	},
	{
		desc: "renewable orphan token",
		req:  &vaultapi.TokenCreateRequest{Lease: "1m", TTL: "1m", NoParent: true, Policies: []string{"fabio"}},
	},
	{
		desc:     "non-renewable orphan token",
		req:      &vaultapi.TokenCreateRequest{Lease: "1m", TTL: "1m", NoParent: true, Renewable: new(bool), Policies: []string{"fabio"}},
		dropWarn: true,
	},
	{
		desc:    "renewable wrapped token",
		wrapTTL: "10s",
		req:     &vaultapi.TokenCreateRequest{Lease: "1m", TTL: "1m", Policies: []string{"fabio"}},
	},
	{
		desc:     "non-renewable wrapped token",
		wrapTTL:  "10s",
		req:      &vaultapi.TokenCreateRequest{Lease: "1m", TTL: "1m", Renewable: new(bool), Policies: []string{"fabio"}},
		dropWarn: true,
	},
}

func TestVaultSource(t *testing.T) {
	const (
		addr      = "127.0.0.1:58421"
		rootToken = "token"
		certPath  = "secret/fabio/cert"
	)

	// start a vault server
	vault, client := vaultServer(t, addr, rootToken)
	defer vault.Process.Kill()

	// create a cert and store it in vault
	certPEM, keyPEM := makePEM("localhost", time.Minute)
	data := map[string]interface{}{"cert": string(certPEM), "key": string(keyPEM)}

	var nilSource *VaultSource // for calling helper methods

	mountPath, v2, err := nilSource.isKVv2(certPath, client)
	if err != nil {
		t.Fatal(err)
	}

	p := certPath + "/localhost"
	if v2 {
		t.Log("Vault: KV backend: V2")
		data = map[string]interface{}{
			"data":    data,
			"options": map[string]interface{}{},
		}
		p = nilSource.addPrefixToVKVPath(p, mountPath, "data")
	} else {
		t.Log("Vault: KV backend: V1")
	}
	if _, err := client.Logical().Write(p, data); err != nil {
		t.Fatalf("logical.Write failed: %s", err)
	}

	pool := makeCertPool(certPEM)
	timeout := 500 * time.Millisecond
	for _, tt := range vaultTestCases {
		tt := tt // capture loop var
		t.Run(tt.desc, func(t *testing.T) {
			src := &VaultSource{
				Client: &vaultClient{
					addr:  "http://" + addr,
					token: makeToken(t, client, tt.wrapTTL, tt.req),
				},
				CertPath: certPath,
			}

			// suppress the log warning about a non-renewable token
			// since this is the expected behavior.
			dropNotRenewableWarning = tt.dropWarn
			testSource(t, src, pool, timeout)
			dropNotRenewableWarning = false
		})
	}
}

func TestVaultPKISource(t *testing.T) {
	const (
		addr      = "127.0.0.1:58421"
		rootToken = "token"
		certPath  = "test-pki/issue/fabio"
	)

	// start a vault server
	vault, client := vaultServer(t, addr, rootToken)
	defer vault.Process.Kill()

	// mount the PKI backend
	err := client.Sys().Mount("test-pki", &vaultapi.MountInput{
		Type: "pki",
		Config: vaultapi.MountConfigInput{
			DefaultLeaseTTL: "1h", // default validity period of issued certificates
			MaxLeaseTTL:     "2h", // maximum validity period of issued certificates
		},
	})
	if err != nil {
		t.Fatalf("Mount pki backend failed: %s", err)
	}

	// generate root CA cert
	resp, err := client.Logical().Write("test-pki/root/generate/internal", map[string]interface{}{
		"common_name": "fabio-ca.com",
		"ttl":         "2h",
	})
	if err != nil {
		t.Fatalf("Generate root failed: %s", err)
	}
	caPool := makeCertPool([]byte(resp.Data["certificate"].(string)))

	// create role
	role := filepath.Base(certPath)
	_, err = client.Logical().Write("test-pki/roles/"+role, map[string]interface{}{
		"allowed_domains": "",
		"allow_localhost": true,
		"allow_ip_sans":   true,
		"organization":    "Fabio Test",
	})
	if err != nil {
		t.Fatalf("Write role failed: %s", err)
	}

	for _, tt := range vaultTestCases {
		tt := tt // capture loop var
		t.Run(tt.desc, func(t *testing.T) {
			src := NewVaultPKISource()
			src.Client = &vaultClient{
				addr:  "http://" + addr,
				token: makeToken(t, client, tt.wrapTTL, tt.req),
			}
			src.CertPath = certPath

			// suppress the log warning about a non-renewable token
			// since this is the expected behavior.
			dropNotRenewableWarning = tt.dropWarn
			testSource(t, src, caPool, 0)
			dropNotRenewableWarning = false
		})
	}
}

// testSource runs an integration test by making an HTTPS request
// to https://localhost/ expecting that the source provides a valid
// certificate for "localhost". rootCAs is expected to contain a
// valid root certificate or the server certificate itself so that
// the HTTPS client can validate the certificate presented by the
// server.
func testSource(t *testing.T, source Source, rootCAs *x509.CertPool, sleep time.Duration) {
	const NoStrictMatch = false
	srvConfig, err := TLSConfig(source, NoStrictMatch, 0, 0, nil)
	if err != nil {
		t.Fatalf("TLSConfig: got %q want nil", err)
	}

	// give the source some time to initialize if necessary
	time.Sleep(sleep)

	// create an http client that will accept the root CAs
	// otherwise the HTTPS client will not verify the
	// certificate presented by the server.
	http11 := http11Client(rootCAs)
	http20, err := http20Client(rootCAs)
	if err != nil {
		t.Fatal("http20Client: ", err)
	}

	// disable log output for the next call to prevent
	// confusing log messages since they are expected
	// http: TLS handshake error from 127.0.0.1:55044: remote error: bad certificate
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// fail calls https://localhost.org/ for which certificate validation
	// should fail since the hostname differs from the one in the certificate.
	fail := func(client *http.Client) {
		_, _, err := roundtrip("localhost.org", srvConfig, client)
		got, want := err, "x509: certificate is valid for localhost, not localhost.org"
		if got == nil || !strings.Contains(got.Error(), want) {
			t.Fatalf("got %q want %q", got, want)
		}
	}

	// succeed executes a roundtrip to https://localhost/ which
	// should return 200 OK and wantBody.
	succeed := func(client *http.Client, wantBody string) {
		code, body, err := roundtrip("localhost", srvConfig, client)
		if err != nil {
			t.Fatalf("got %v want nil", err)
		}
		if got, want := code, 200; got != want {
			t.Fatalf("got %v want %v", got, want)
		}
		if got, want := body, wantBody; got != want {
			t.Fatalf("got %v want %v", got, want)
		}
	}

	// make a call for which certificate validation succeeds.
	succeed(http11, "OK HTTP/1.1")
	succeed(http20, "OK HTTP/2.0")

	// now make the call that should fail.
	fail(http11)
	fail(http20)
}

// roundtrip starts a TLS server with the given server configuration and
// then sends an SNI request with the given serverName.
func roundtrip(serverName string, srvConfig *tls.Config, client *http.Client) (code int, body string, err error) {
	// create an HTTPS server and start it. It will be listening on 127.0.0.1
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK ", r.Proto)
	}))
	srv.TLS = srvConfig
	srv.StartTLS()
	defer srv.Close()

	// configure SNI
	client.Transport.(*http.Transport).TLSClientConfig.ServerName = serverName

	// give the tls server some time to start up
	time.Sleep(10 * time.Millisecond)

	resp, err := client.Get(srv.URL)
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

// http11Client returns an HTTP client which can only
// execute HTTP/1.1 requests via TLS.
func http11Client(rootCAs *x509.CertPool) *http.Client {
	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: rootCAs,
		},
	}
	return &http.Client{Transport: t}
}

// http20Client returns an HTTP client which can
// execute HTTP/2.0 requests via TLS if the server
// supports it.
func http20Client(rootCAs *x509.CertPool) (*http.Client, error) {
	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: rootCAs,
		},
	}
	if err := http2.ConfigureTransport(t); err != nil {
		return nil, err
	}
	return &http.Client{Transport: t}, nil
}

func tempDir() string {
	dir, err := ioutil.TempDir("", "fabio")
	if err != nil {
		panic(err.Error())
	}
	return dir
}

func writeFile(filename string, data []byte) {
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		panic(err.Error())
	}
}

func makeCertPool(x ...[]byte) *x509.CertPool {
	p := x509.NewCertPool()
	for _, b := range x {
		// https://github.com/fabiolb/fabio/issues/434
		if ok := p.AppendCertsFromPEM(b); !ok {
			panic("failed to add cert from PEM. Is the CN a DNS compatible name?")
		}
	}
	return p
}

func saveCert(dir, host string, certPEM, keyPEM []byte) (certFile, keyFile string) {
	certFile, keyFile = filepath.Join(dir, host+"-cert.pem"), filepath.Join(dir, host+"-key.pem")
	writeFile(certFile, certPEM)
	writeFile(keyFile, keyPEM)
	return certFile, keyFile
}

// makePEM creates a self-signed RSA certificate as two PEM blocks.
// taken from crypto/tls/generate_cert.go
func makePEM(host string, validFor time.Duration) (certPEM, keyPEM []byte) {
	const bits = 1024
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		panic("Failed to generate private key: " + err.Error())
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
		panic("Failed to create certificate: " + err.Error())
	}

	var cert, key bytes.Buffer
	pem.Encode(&cert, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	pem.Encode(&key, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return cert.Bytes(), key.Bytes()
}

func makeCert(host string, validFor time.Duration) tls.Certificate {
	certPEM, keyPEM := makePEM(host, validFor)
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic("Failed to create certificate: " + err.Error())
	}
	return cert
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
