package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fabiolb/fabio/config"
)

func TestAdminServerAccess(t *testing.T) {
	type test struct {
		uri  string
		code int
	}

	testAccess := func(access string, tests []test) {
		srv := &Server{
			Access: access,
			Cfg: &config.Config{
				Registry: config.Registry{
					Consul: config.Consul{
						KVPath: "/fabio/config",
					},
				},
			},
		}
		ts := httptest.NewServer(srv.handler())
		defer ts.Close()

		noRedirectClient := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		for _, tt := range tests {
			t.Run(access+tt.uri, func(t *testing.T) {
				resp, err := noRedirectClient.Get(ts.URL + tt.uri)
				if err != nil {
					t.Fatalf("got %v want nil", err)
				}
				if got, want := resp.StatusCode, tt.code; got != want {
					t.Fatalf("got code %d want %d", got, want)
				}
			})
		}
	}

	roTests := []test{
		{"/api/manual", 403},
		{"/api/paths", 403},
		{"/api/config", 200},
		{"/api/routes", 200},
		{"/api/version", 200},
		{"/manual", 403},
		{"/routes", 200},
		{"/health", 200},
		{"/assets/logo.svg", 200},
		{"/assets/logo.bw.svg", 200},
		{"/", 303},
	}

	rwTests := []test{
		{"/api/manual", 200},
		{"/api/paths", 200},
		{"/api/config", 200},
		{"/api/routes", 200},
		{"/api/version", 200},
		{"/manual", 200},
		{"/routes", 200},
		{"/health", 200},
		{"/assets/logo.svg", 200},
		{"/assets/logo.bw.svg", 200},
		{"/", 303},
	}

	testAccess("ro", roTests)
	testAccess("rw", rwTests)
}
