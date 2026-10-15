package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/fabiolb/fabio/route"
)

type RoutesHandler struct{}

type apiRoute struct {
	Service string   `json:"service"`
	Host    string   `json:"host"`
	Path    string   `json:"path"`
	Src     string   `json:"src"`
	Dst     string   `json:"dst"`
	Opts    string   `json:"opts"`
	Cmd     string   `json:"cmd"`
	Tags    []string `json:"tags,omitempty"`
	Weight  float64  `json:"weight"`
	Rate1   float64  `json:"rate1"`
	Pct99   float64  `json:"pct99"`
}

func (h *RoutesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := route.GetTable()

	if _, ok := r.URL.Query()["raw"]; ok {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, t.String())
		return
	}

	var hosts []string
	for host := range t {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)

	var routes []apiRoute
	for _, host := range hosts {
		for _, tr := range t[host] {
			for _, tg := range tr.Targets {
				routes = append(routes, toAPIRoute(tr, tg, "route add", tg.Weight))
			}
			for _, tg := range tr.DeletedTargets {
				routes = append(routes, toAPIRoute(tr, tg, "route del", 0))
			}
		}
	}
	writeJSON(w, r, routes)
}

// toAPIRoute presents live and removed targets with the same route metadata.
func toAPIRoute(tr *route.Route, tg *route.Target, cmd string, weight float64) apiRoute {
	var opts []string
	for k, v := range tg.Opts {
		opts = append(opts, k+"="+v)
	}

	return apiRoute{
		Service: tg.Service,
		Host:    tr.Host,
		Path:    tr.Path,
		Src:     tr.Host + tr.Path,
		Dst:     tg.URL.String(),
		Opts:    strings.Join(opts, " "),
		Weight:  weight,
		Tags:    tg.Tags,
		Cmd:     cmd,
	}
}
