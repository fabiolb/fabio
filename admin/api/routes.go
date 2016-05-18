package api

import (
	"fmt"
	"net/http"
	"sort"

	fabioroute "github.com/eBay/fabio/route"
)

type route struct {
	Service string   `json:"service"`
	Host    string   `json:"host"`
	Path    string   `json:"path"`
	Dst     string   `json:"dst"`
	Weight  float64  `json:"weight"`
	Tags    []string `json:"tags,omitempty"`
	Cmd     string   `json:"cmd"`
	Rate1   float64  `json:"rate1"`
	Pct99   float64  `json:"pct99"`
}

// HandleRoutes provides a fetch handler for the current routing table.
func HandleRoutes(w http.ResponseWriter, r *http.Request) {
	t := fabioroute.GetTable()

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

	var routes []route
	for _, host := range hosts {
		for _, tr := range t[host] {
			for _, tg := range tr.Targets {
				ar := route{
					Service: tg.Service,
					Host:    tr.Host,
					Path:    tr.Path,
					Dst:     tg.URL.String(),
					Weight:  tg.Weight,
					Tags:    tg.Tags,
					Cmd:     tr.TargetConfig(tg, true),
					Rate1:   tg.Timer.Rate1(),
					Pct99:   tg.Timer.Percentile(0.99),
				}
				routes = append(routes, ar)
			}
		}
	}
	writeJSON(w, r, routes)
}
