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
	Weight  float64  `json:"weight"`
	Tags    []string `json:"tags,omitempty"`
	Cmd     string   `json:"cmd"`
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
				var opts []string
				for k, v := range tg.Opts {
					opts = append(opts, k+"="+v)
				}

				ar := apiRoute{
					Service: tg.Service,
					Host:    tr.Host,
					Path:    tr.Path,
					Src:     tr.Host + tr.Path,
					Dst:     tg.URL.String(),
					Opts:    strings.Join(opts, " "),
					Weight:  tg.Weight,
					Tags:    tg.Tags,
					Cmd:     "route add",
					Rate1:   tg.Timer.Rate1(),
					Pct99:   tg.Timer.Percentile(0.99),
				}
				routes = append(routes, ar)
			}
		}
	}
	writeJSON(w, r, routes)
}
