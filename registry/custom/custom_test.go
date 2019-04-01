package custom

import (
	"encoding/json"
	"fmt"
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/route"
	"net/http"
	"testing"
	"time"
)

func TestCustomRoutes(t *testing.T) {

	var resp string
	cfg := config.Custom{
		Host:               "localhost:8080",
		Path:               "test",
		Scheme:             "http",
		CheckTLSSkipVerify: false,
		PollingInterval:    3 * time.Second,
		Timeout:            3 * time.Second,
	}

	ch := make(chan string, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/test", handleTest)
	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
	go server.ListenAndServe()
	time.Sleep(3 * time.Second)
	defer server.Close()

	go customRoutes(&cfg, ch)

	resp = <-ch

	if resp != "OK" {
		fmt.Printf("Failed to get routes for custom backend - %s", resp)
		t.FailNow()
	}

	return

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
