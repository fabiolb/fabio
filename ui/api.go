package ui

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/eBay/fabio/route"
)

type apiRoute struct {
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

func handleRoutes(w http.ResponseWriter, r *http.Request) {
	t := route.GetTable()

	var hosts []string
	for host := range t {
		hosts = append(hosts, host)
	}

	var apiRoutes []apiRoute
	for _, host := range hosts {
		for _, tr := range t[host] {
			for _, tg := range tr.Targets {
				ar := apiRoute{
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
				apiRoutes = append(apiRoutes, ar)
			}
		}
	}
	writeJSON(w, r, apiRoutes)
}

func writeJSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	_, pretty := r.URL.Query()["pretty"]

	var buf []byte
	var err error
	if pretty {
		buf, err = json.MarshalIndent(v, "", "    ")
	} else {
		buf, err = json.Marshal(v)
	}

	if err != nil {
		log.Printf("[ERROR] ", err)
		http.Error(w, "internal error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(buf)
}
