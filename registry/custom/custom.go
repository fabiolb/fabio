package custom

import (
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/route"
	"time"
)

func customRoutes(cfg *config.CustomBE, ch chan string) () {

	var Routes []*route.RouteDef

	for {


		//TODO Get data from BE


		t, err := route.NewTableCustomBE(Routes)
		if err != nil {
			ch <- err.Error()
		}
		route.SetTable(t)

		ch <- "OK"
		time.Sleep(cfg.PollingInterval *time.Second)

	}


}

