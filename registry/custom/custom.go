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

func customRoutes(cfg *config.Custom, ch chan string) {

	var Routes *[]route.RouteDef
	var trans *http.Transport
	var URL string

	if !cfg.CheckTLSSkipVerify {
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
	} else {
		URL = fmt.Sprintf("%s://%s/%s", cfg.Scheme, cfg.Host, cfg.Path)
	}

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Printf("[ERROR] Can not generate new HTTP request")
	}
	req.Close = true

	for {
		log.Printf("[DEBUG] Custom Registry starting request %s \n", time.Now())
		resp, err := client.Do(req)
		if err != nil {
			ch <- fmt.Sprintf("Error Sending HTTPs Request To Custom be - %s -%s", URL, err.Error())
			time.Sleep(cfg.PollingInterval)
			continue
		}

		if resp.StatusCode != 200 {
			ch <- fmt.Sprintf("Error Non-200 return (%v) from  -%s", resp.StatusCode, URL)
			time.Sleep(cfg.PollingInterval)
			continue
		}
		log.Printf("[DEBUG] Custom Registry begin decoding json %s \n", time.Now())
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&Routes)
		if err != nil {
			ch <- fmt.Sprintf("Error decoding request - %s -%s", URL, err.Error())
			time.Sleep(cfg.PollingInterval)
			continue
		}

		log.Printf("[DEBUG] Custom Registry building table %s \n", time.Now())
		t, err := route.NewTableCustom(Routes)
		if err != nil {
			ch <- fmt.Sprintf("Error generating new table - %s", err.Error())
		}
		log.Printf("[DEBUG] Custom Registry building table complete %s \n", time.Now())
		route.SetTable(t)
		log.Printf("[DEBUG] Custom Registry table set complete %s \n", time.Now())
		ch <- "OK"
		time.Sleep(cfg.PollingInterval)

	}

}
