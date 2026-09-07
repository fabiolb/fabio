package custom

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/route"
)

func TestCustomRoutes(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", handleTest)
	server := httptest.NewServer(mux)
	defer server.Close()

	host, _ := strings.CutPrefix(server.URL, "http://")
	cfg := config.Custom{
		Host:               host,
		Path:               "test",
		Scheme:             "http",
		CheckTLSSkipVerify: false,
		PollInterval:       3 * time.Second,
		Timeout:            3 * time.Second,
	}
	ch := make(chan string, 1)

	go customRoutes(&cfg, ch)

	resp := <-ch

	if resp != "OK" {
		t.Fatalf("Failed to get routes for custom backend - %s", resp)
	}
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	var routes []route.RouteDef
	var tags = []string{"tag1", "tag2"}
	var opts = make(map[string]string)
	opts["tlsskipverify"] = "true"
	opts["proto"] = "http"

	var route1 = route.RouteDef{
		Cmd:     "route add",
		Service: "service1",
		Src:     "app.com",
		Dst:     "http://10.1.1.1:8080",
		Weight:  0.50,
		Tags:    tags,
		Opts:    opts,
	}

	var route2 = route.RouteDef{
		Cmd:     "route add",
		Service: "service1",
		Src:     "app.com",
		Dst:     "http://10.1.1.2:8080",
		Weight:  0.50,
		Tags:    tags,
		Opts:    opts,
	}
	var route3 = route.RouteDef{
		Cmd:     "route add",
		Service: "service2",
		Src:     "app.com",
		Dst:     "http://10.1.1.3:8080",
		Weight:  0.25,
		Tags:    tags,
		Opts:    opts,
	}

	routes = append(routes, route1)
	routes = append(routes, route2)
	routes = append(routes, route3)

	rt, _ := json.Marshal(routes)

	w.Write(rt)
}
