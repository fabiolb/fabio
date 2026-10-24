package consul

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
)

func consulTestCA(t *testing.T, serial int64) (*x509.Certificate, *ecdsa.PrivateKey, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(serial),
		Subject:               pkix.Name{CommonName: "consul-test-ca"},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return cert, key, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func consulTestCertificate(t *testing.T, ca *x509.Certificate, signer *ecdsa.PrivateKey, serial int64, commonName string, server bool) (tls.Certificate, []byte, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	if server {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	}
	der, err := x509.CreateCertificate(rand.Reader, template, ca, &key.PublicKey, signer)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	return pair, certPEM, keyPEM
}

func replaceConsulTLSFile(t *testing.T, dest string, data []byte) {
	t.Helper()
	file, err := os.CreateTemp(filepath.Dir(dest), "consul-tls-")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.Write(data); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(file.Name(), dest); err != nil {
		t.Fatal(err)
	}
}

func TestConsulTLSReload(t *testing.T) {
	ca1, caKey1, caPEM1 := consulTestCA(t, 1)
	ca2, caKey2, caPEM2 := consulTestCA(t, 2)
	serverCert1, _, _ := consulTestCertificate(t, ca1, caKey1, 3, "consul-server-1", true)
	serverCert2, _, _ := consulTestCertificate(t, ca2, caKey2, 4, "consul-server-2", true)
	_, certPEM1, keyPEM1 := consulTestCertificate(t, ca1, caKey1, 5, "client-1", false)
	_, certPEM2, keyPEM2 := consulTestCertificate(t, ca2, caKey2, 6, "client-2", false)

	dir := t.TempDir()
	certFile := filepath.Join(dir, "client.crt")
	keyFile := filepath.Join(dir, "client.key")
	caFile := filepath.Join(dir, "ca.pem")
	replaceConsulTLSFile(t, certFile, certPEM1)
	replaceConsulTLSFile(t, keyFile, keyPEM1)
	replaceConsulTLSFile(t, caFile, caPEM1)

	clientCAs := x509.NewCertPool()
	clientCAs.AppendCertsFromPEM(caPEM1)
	clientCAs.AppendCertsFromPEM(caPEM2)
	serverConfig := func(cert tls.Certificate) *tls.Config {
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clientCAs,
		}
	}
	var currentServer atomic.Pointer[tls.Config]
	currentServer.Store(serverConfig(serverCert1))
	var closeConnections atomic.Bool

	var mu sync.Mutex
	var lastClient string
	entered := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		lastClient = r.TLS.PeerCertificates[0].Subject.CommonName
		mu.Unlock()
		if closeConnections.Load() {
			w.Header().Set("Connection", "close")
		}
		if r.URL.Path == "/v1/kv/slow" {
			close(entered)
			<-release
		}
		if r.URL.Path == "/v1/agent/self" {
			json.NewEncoder(w).Encode(map[string]any{"Config": map[string]any{"Datacenter": "test"}})
			return
		}
		json.NewEncoder(w).Encode([]any{})
	}))
	server.TLS = &tls.Config{
		GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
			return currentServer.Load(), nil
		},
	}
	server.StartTLS()
	defer func() {
		releaseOnce.Do(func() { close(release) })
		server.Close()
	}()

	cfg := &config.Consul{
		Scheme: "https",
		Addr:   server.Listener.Addr().String(),
		TLS: config.ConsulTlS{
			CertFile: certFile,
			KeyFile:  keyFile,
			CAFile:   caFile,
		},
	}
	backend, err := NewBackend(cfg)
	if err != nil {
		t.Fatal(err)
	}
	b := backend.(*be)
	client := b.c
	checkClient := func(want string) {
		t.Helper()
		if _, err := b.ManualPaths(); err != nil {
			t.Fatal(err)
		}
		mu.Lock()
		got := lastClient
		mu.Unlock()
		if got != want {
			t.Fatalf("client certificate = %q, want %q", got, want)
		}
		if b.c != client {
			t.Fatal("Consul API client was replaced")
		}
	}
	checkClient("client-1")

	// Keep an old request in flight while the next request changes transports.
	slowDone := make(chan error, 1)
	go func() {
		_, _, err := client.KV().List("slow", nil)
		slowDone <- err
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("old request did not reach Consul")
	}

	// A half-updated pair must continue to use the previously validated TLS config.
	replaceConsulTLSFile(t, certFile, certPEM2)
	checkClient("client-1")
	replaceConsulTLSFile(t, keyFile, keyPEM2)
	checkClient("client-2")
	releaseOnce.Do(func() { close(release) })
	select {
	case err := <-slowDone:
		if err != nil {
			t.Fatalf("in-flight request failed during rotation: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("in-flight request did not finish")
	}

	// Failed reads and mismatched keys must leave the working client in place.
	// Force fresh handshakes to check the retained TLS configuration.
	closeConnections.Store(true)
	checkClient("client-2")
	if err := os.Remove(certFile); err != nil {
		t.Fatal(err)
	}
	checkClient("client-2")
	replaceConsulTLSFile(t, certFile, certPEM1)
	checkClient("client-2")
	replaceConsulTLSFile(t, keyFile, keyPEM1)
	checkClient("client-1")

	// Rotate the server's CA, requiring the next connection to trust the new CA.
	currentServer.Store(serverConfig(serverCert2))
	replaceConsulTLSFile(t, caFile, caPEM2)
	checkClient("client-1")
	replaceConsulTLSFile(t, caFile, []byte("not a certificate"))
	checkClient("client-1")
	replaceConsulTLSFile(t, caFile, caPEM2)
	checkClient("client-1")
}

func TestConsulTLSReloadConfiguration(t *testing.T) {
	// No TLS file paths means the original plain HTTP and HTTPS paths remain in use.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/agent/self" {
			json.NewEncoder(w).Encode(map[string]any{"Config": map[string]any{"Datacenter": "test"}})
			return
		}
		json.NewEncoder(w).Encode([]any{})
	})
	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()
	httpsServer := httptest.NewTLSServer(handler)
	defer httpsServer.Close()
	caDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(caDir, "server.pem"), pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE", Bytes: httpsServer.Certificate().Raw,
	}), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name string
		cfg  config.Consul
	}{
		{name: "http", cfg: config.Consul{Addr: httpServer.Listener.Addr().String(), Scheme: "http"}},
		{name: "https without certificate paths", cfg: config.Consul{
			Addr:   httpsServer.Listener.Addr().String(),
			Scheme: "https",
			TLS:    config.ConsulTlS{InsecureSkipVerify: true},
		}},
		{name: "https with CAPath", cfg: config.Consul{
			Addr:   httpsServer.Listener.Addr().String(),
			Scheme: "https",
			TLS:    config.ConsulTlS{CAPath: caDir},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := NewBackend(&tt.cfg)
			if err != nil {
				t.Fatal(err)
			}
			b := backend.(*be)
			if _, err := b.ManualPaths(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestConsulTLSReloadInvalidInitialPair(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "client.crt")
	keyFile := filepath.Join(dir, "client.key")
	ca, caKey, _ := consulTestCA(t, 10)
	_, _, keyPEM1 := consulTestCertificate(t, ca, caKey, 11, "client-1", false)
	_, certPEM2, _ := consulTestCertificate(t, ca, caKey, 12, "client-2", false)
	replaceConsulTLSFile(t, certFile, certPEM2)
	replaceConsulTLSFile(t, keyFile, keyPEM1)
	_, err := NewBackend(&config.Consul{
		Scheme: "https",
		TLS: config.ConsulTlS{
			CertFile: certFile,
			KeyFile:  keyFile,
		},
	})
	if err == nil || !strings.Contains(err.Error(), "private key does not match public key") {
		t.Fatalf("mismatched initial certificate and key must fail TLS setup: %v", err)
	}
}
