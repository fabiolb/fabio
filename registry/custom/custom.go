package custom

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/route"
	"log"
	"net/http"
	"time"
)

func customRoutes(cfg *config.CustomBE, ch chan string) {

	var Routes *[]route.RouteDef
	var trans *http.Transport
	var URL string

	if cfg.CheckTLSSkipVerify {
		trans = &http.Transport{}

	} else {
		trans = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client := &http.Client{
		Transport: trans,
		Timeout:   cfg.Timeout,
	}

	if cfg.QueryParams != "" {
		URL = fmt.Sprintf("%s://%s/%s?%s", cfg.Scheme, cfg.Host, cfg.Path, cfg.QueryParams)
	}else {
		URL = fmt.Sprintf("%s://%s/%s", cfg.Scheme, cfg.Host, cfg.Path)
	}

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Printf("[ERROR] Can not generate new HTTP request")
	}
	req.Close = true


	for {

		resp, err := client.Do(req)
		if err != nil {
			ch <- fmt.Sprintf("Error Sending HTTPs Request To Custom BE - %s -%s", URL, err.Error())
			time.Sleep(cfg.PollingInterval)
			continue
		}

		if resp.StatusCode != 200 {
			ch <- fmt.Sprintf("Error Non-200 return (%v) from  -%s", resp.StatusCode, URL)
			time.Sleep(cfg.PollingInterval)
			continue
		}

		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&Routes)
		if err != nil {
			ch <- fmt.Sprintf("Error decoding request - %s -%s", URL, err.Error())
			time.Sleep(cfg.PollingInterval)
			continue
		}

		//TODO validate data

		t, err := route.NewTableCustomBE(Routes)
		if err != nil {
			ch <- fmt.Sprintf("Error generating new table - %s", err.Error())
		}
		route.SetTable(t)

		ch <- "OK"
		fmt.Printf("got from %s\n", URL)
		time.Sleep(cfg.PollingInterval)

	}

}
