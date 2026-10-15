package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/route"
)

func TestAdminServerAccess(t *testing.T) {
	type test struct {
		uri  string
		code int
	}

	testAccess := func(access, basePath string, tests []test) {
		srv := &Server{
			Access: access,
			Path:   basePath,
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

	noTests := []test{
		{"/api/manual", 404},
		{"/api/paths", 404},
		{"/api/config", 404},
		{"/api/routes", 404},
		{"/api/version", 404},
		{"/manual", 404},
		{"/routes", 404},
		{"/health", 200},
		{"/assets/logo.svg", 404},
		{"/assets/logo.bw.svg", 404},
		{"/", 404},
	}

	testAccess("ro", "", roTests)
	testAccess("rw", "", rwTests)
	testAccess("no", "", noTests)

	roTestsWithPath := []test{
		{"/fabio/api/manual", 403},
		{"/fabio/api/paths", 403},
		{"/fabio/api/config", 200},
		{"/fabio/api/routes", 200},
		{"/fabio/api/version", 200},
		{"/fabio/manual", 403},
		{"/fabio/routes", 200},
		{"/health", 200},
		{"/fabio/assets/logo.svg", 200},
		{"/fabio/assets/logo.bw.svg", 200},
		{"/fabio/", 303},
	}

	rwTestsWithPath := []test{
		{"/fabio/api/manual", 200},
		{"/fabio/api/paths", 200},
		{"/fabio/api/config", 200},
		{"/fabio/api/routes", 200},
		{"/fabio/api/version", 200},
		{"/fabio/manual", 200},
		{"/fabio/routes", 200},
		{"/health", 200},
		{"/fabio/assets/logo.svg", 200},
		{"/fabio/assets/logo.bw.svg", 200},
		{"/fabio/", 303},
	}

	noTestsWithPath := []test{
		{"/fabio/api/manual", 404},
		{"/fabio/api/paths", 404},
		{"/fabio/api/config", 404},
		{"/fabio/api/routes", 404},
		{"/fabio/api/version", 404},
		{"/fabio/manual", 404},
		{"/fabio/routes", 404},
		{"/health", 200},
		{"/fabio/assets/logo.svg", 404},
		{"/fabio/assets/logo.bw.svg", 404},
		{"/fabio/", 404},
	}

	testAccess("ro", "/fabio", roTestsWithPath)
	testAccess("rw", "/fabio", rwTestsWithPath)
	testAccess("no", "/fabio", noTestsWithPath)
}

func TestDeletedRouteAPIAndPageForReadOnlyAndWritableAccess(t *testing.T) {
	old := route.GetTable()
	defer route.SetTable(old)

	table, err := route.NewTable(bytes.NewBufferString(`route add removed /old http://gone.example/ tags "retired" opts "strip=/old"
route add live /live http://live.example/
route add zero /zero http://zero.example/ weight 0
route del removed`))
	if err != nil {
		t.Fatal(err)
	}
	route.SetTable(table)

	for _, access := range []string{"ro", "rw"} {
		t.Run(access, func(t *testing.T) {
			srv := &Server{Access: access, Cfg: &config.Config{}}
			request := httptest.NewRequest("GET", "/api/routes", nil)
			rec := httptest.NewRecorder()
			srv.handler().ServeHTTP(rec, request)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d", rec.Code)
			}
			var rows []struct {
				Service, Src, Dst, Cmd, Opts string
				Tags                         []string
				Weight                       float64
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &rows); err != nil {
				t.Fatal(err)
			}
			if len(rows) != 3 {
				t.Fatalf("rows = %+v", rows)
			}
			for _, row := range rows {
				switch row.Service {
				case "removed":
					if row.Cmd != "route del" || row.Weight != 0 || row.Src != "/old" || row.Dst != "http://gone.example/" || row.Opts != "strip=/old" || len(row.Tags) != 1 || row.Tags[0] != "retired" {
						t.Errorf("deleted row = %+v", row)
					}
				case "live":
					if row.Cmd != "route add" || row.Weight != 1 || row.Dst != "http://live.example/" {
						t.Errorf("live row = %+v", row)
					}
				case "zero":
					if row.Cmd != "route add" || row.Weight != 0 || row.Dst != "http://zero.example/" {
						t.Errorf("zero-weight active row = %+v", row)
					}
				default:
					t.Errorf("unexpected row = %+v", row)
				}
			}

			raw := httptest.NewRecorder()
			srv.handler().ServeHTTP(raw, httptest.NewRequest("GET", "/api/routes?raw", nil))
			if raw.Code != http.StatusOK || strings.Contains(raw.Body.String(), "gone.example") || strings.Contains(raw.Body.String(), "route del") || !strings.Contains(raw.Body.String(), "live.example") {
				t.Errorf("raw route config = %q (status %d)", raw.Body.String(), raw.Code)
			}

			page := httptest.NewRecorder()
			srv.handler().ServeHTTP(page, httptest.NewRequest("GET", "/routes", nil))
			if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), `r.cmd === "route del"`) || !strings.Contains(page.Body.String(), "tr.deleted-route td") {
				t.Errorf("routes page missing deleted state rendering (status %d)", page.Code)
			}
		})
	}
}
