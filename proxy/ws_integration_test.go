package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/route"

	"golang.org/x/net/websocket"
)

func TestProxyWSUpstream(t *testing.T) {
	wsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/ws", "/wss", "/insecure", "/strip", "/foo/bar/baz", "/new/world":
			websocket.Handler(wsEchoHandler).ServeHTTP(w, r)
		default:
			w.WriteHeader(404)
		}
	}))
	defer wsServer.Close()
	t.Log("Started WS server: ", wsServer.URL)

	wssServer := httptest.NewUnstartedServer(websocket.Handler(wsEchoHandler))
	wssServer.TLS = tlsServerConfig()
	wssServer.StartTLS()
	defer wssServer.Close()
	t.Log("Started WSS server: ", wssServer.URL)

	routes := "route add ws /ws  " + wsServer.URL + "\n"
	routes += "route add ws /wss " + wssServer.URL + ` opts "proto=https"` + "\n"
	routes += "route add ws /insecure " + wssServer.URL + ` opts "proto=https tlsskipverify=true"` + "\n"
	routes += "route add ws /foo/strip  " + wsServer.URL + ` opts "strip=/foo"` + "\n"
	routes += "route add ws /bar/baz  " + wsServer.URL + ` opts "prepend=/foo"` + "\n"
	routes += "route add ws /old/world  " + wsServer.URL + ` opts "prepend=/new strip=/old"` + "\n"

	httpProxy := httptest.NewServer(&HTTPProxy{
		Config:            config.Proxy{NoRouteStatus: 404, GZIPContentTypes: regexp.MustCompile(".*")},
		Transport:         &http.Transport{TLSClientConfig: tlsClientConfig()},
		InsecureTransport: &http.Transport{TLSClientConfig: tlsInsecureConfig()},
		Lookup: func(r *http.Request) *route.Target {
			tbl, _ := route.NewTable(bytes.NewBufferString(routes))
			return tbl.Lookup(r, route.Picker["rr"], route.Matcher["prefix"], globCache, globEnabled)
		},
	})
	defer httpProxy.Close()
	t.Log("Started HTTP proxy: ", httpProxy.URL)

	httpsProxy := httptest.NewUnstartedServer(&HTTPProxy{
		Config:            config.Proxy{NoRouteStatus: 404},
		Transport:         &http.Transport{TLSClientConfig: tlsClientConfig()},
		InsecureTransport: &http.Transport{TLSClientConfig: tlsInsecureConfig()},
		Lookup: func(r *http.Request) *route.Target {
			tbl, _ := route.NewTable(bytes.NewBufferString(routes))
			return tbl.Lookup(r, route.Picker["rr"], route.Matcher["prefix"], globCache, globEnabled)
		},
	})
	httpsProxy.TLS = tlsServerConfig()
	httpsProxy.StartTLS()
	defer httpsProxy.Close()
	t.Log("Started HTTPS proxy: ", httpsProxy.URL)

	wsServerURL := wsServer.URL[len("http://"):]
	wssServerURL := wssServer.URL[len("https://"):]
	httpProxyURL := httpProxy.URL[len("http://"):]
	httpsProxyURL := httpsProxy.URL[len("https://"):]

	t.Run("ws-ws direct", func(t *testing.T) { testWSEcho(t, "ws://"+wsServerURL+"/ws", nil) })
	t.Run("wss-wss direct", func(t *testing.T) { testWSEcho(t, "wss://"+wssServerURL+"/wss", nil) })

	t.Run("ws-ws via http proxy", func(t *testing.T) { testWSEcho(t, "ws://"+httpProxyURL+"/ws", nil) })
	t.Run("wss-ws via https proxy", func(t *testing.T) { testWSEcho(t, "wss://"+httpsProxyURL+"/ws", nil) })

	t.Run("ws-wss via http proxy", func(t *testing.T) { testWSEcho(t, "ws://"+httpProxyURL+"/wss", nil) })
	t.Run("wss-wss via https proxy", func(t *testing.T) { testWSEcho(t, "wss://"+httpsProxyURL+"/wss", nil) })

	t.Run("ws-wss tlsskipverify=true via http proxy", func(t *testing.T) { testWSEcho(t, "ws://"+httpProxyURL+"/insecure", nil) })
	t.Run("wss-wss tlsskipverify=true via https proxy", func(t *testing.T) { testWSEcho(t, "wss://"+httpsProxyURL+"/insecure", nil) })

	h := http.Header{"Accept-Encoding": []string{"gzip"}}
	t.Run("ws-ws via http proxy with gzip", func(t *testing.T) { testWSEcho(t, "ws://"+httpProxyURL+"/ws", h) })

	t.Run("ws-ws via http proxy with strip", func(t *testing.T) { testWSEcho(t, "ws://"+httpProxyURL+"/foo/strip", nil) })

	t.Run("ws-ws via http proxy with prepend", func(t *testing.T) { testWSEcho(t, "ws://"+httpProxyURL+"/bar/baz", nil) })

	t.Run("ws-ws via http proxy with strip & prepend", func(t *testing.T) { testWSEcho(t, "ws://"+httpProxyURL+"/old/world", nil) })
}

func TestProxyWSHandshakeResponseTimeout(t *testing.T) {
	t.Parallel()
	const handshakeDelay = 1250 * time.Millisecond
	delayedEcho := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(handshakeDelay)
		websocket.Handler(wsEchoHandler).ServeHTTP(w, r)
	})

	for _, mode := range []struct {
		name        string
		proxyTLS    bool
		upstreamTLS bool
	}{
		{"http to ws", false, false},
		{"https to ws", true, false},
		{"http to wss", false, true},
		{"https to wss", true, true},
	} {
		t.Run(mode.name, func(t *testing.T) {
			t.Parallel()
			upstream := httptest.NewUnstartedServer(delayedEcho)
			if mode.upstreamTLS {
				upstream.TLS = tlsServerConfig()
				upstream.StartTLS()
			} else {
				upstream.Start()
			}
			defer upstream.Close()

			routes := "route add ws /delayed " + upstream.URL
			if mode.upstreamTLS {
				routes += ` opts "proto=https"`
			}
			table, err := route.NewTable(bytes.NewBufferString(routes + "\n"))
			if err != nil {
				t.Fatal(err)
			}

			for _, tc := range []struct {
				name        string
				timeout     time.Duration
				wantSuccess bool
			}{
				{"no limit", 0, true},
				{"long enough", 3 * time.Second, true},
				{"short limit", 250 * time.Millisecond, false},
			} {
				t.Run(tc.name, func(t *testing.T) {
					p := httptest.NewUnstartedServer(&HTTPProxy{
						Config:            config.Proxy{NoRouteStatus: 404, ResponseHeaderTimeout: tc.timeout},
						Transport:         &http.Transport{TLSClientConfig: tlsClientConfig()},
						InsecureTransport: &http.Transport{TLSClientConfig: tlsInsecureConfig()},
						Lookup: func(r *http.Request) *route.Target {
							return table.Lookup(r, route.Picker["rr"], route.Matcher["prefix"], globCache, globEnabled)
						},
					})
					if mode.proxyTLS {
						p.TLS = tlsServerConfig()
						p.StartTLS()
					} else {
						p.Start()
					}
					defer p.Close()

					wsURL := "ws://" + strings.TrimPrefix(p.URL, "http://") + "/delayed"
					if mode.proxyTLS {
						wsURL = "wss://" + strings.TrimPrefix(p.URL, "https://") + "/delayed"
					}
					if tc.wantSuccess {
						testWSEcho(t, wsURL, nil)
						return
					}

					cfg, err := websocket.NewConfig(wsURL, "http://localhost/")
					if err != nil {
						t.Fatal(err)
					}
					if mode.proxyTLS {
						cfg.TlsConfig = tlsClientConfig()
					}
					started := time.Now()
					ws, err := websocket.DialConfig(cfg)
					if ws != nil {
						ws.Close()
					}
					if err == nil {
						t.Fatal("expected the backend handshake to exceed the configured timeout")
					}
					if elapsed := time.Since(started); elapsed >= handshakeDelay {
						t.Fatalf("handshake failed after %s, expected it to time out before %s", elapsed, handshakeDelay)
					}
				})
			}
		})
	}
}

func testWSEcho(t *testing.T, url string, hdr http.Header) {
	cfg, err := websocket.NewConfig(url, "http://localhost/")
	if err != nil {
		t.Fatal("NewConfig: ", err)
	}
	cfg.Header = hdr
	if strings.HasPrefix(url, "wss://") {
		cfg.TlsConfig = tlsClientConfig()
	}
	ws, err := websocket.DialConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	send := []byte("foo")
	if _, err := ws.Write([]byte("foo")); err != nil {
		t.Logf("ws.Write failed: %s", err)
	}
	recv := make([]byte, 100)
	n, err := ws.Read(recv)
	if err != nil {
		t.Logf("ws.Read failed: %s", err)
	}
	recv = recv[:n]
	if got, want := recv, send; !bytes.Equal(got, want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func wsEchoHandler(ws *websocket.Conn) {
	io.Copy(ws, ws)
}
