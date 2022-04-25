// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Broker API support - Fetch and Search
// See: https://login.circonus.com/resources/api/calls/broker

package apiclient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/circonus-labs/go-apiclient/config"
	"github.com/pkg/errors"
)

// BrokerDetail defines instance attributes
type BrokerDetail struct {
	CN           string   `json:"cn"`                       // string
	Status       string   `json:"status"`                   // string
	ClusterIP    *string  `json:"cluster_ip"`               // string or null
	ExternalHost *string  `json:"external_host"`            // string or null
	IP           *string  `json:"ipaddress"`                // string or null
	Skew         *string  `json:"skew"`                     // BUG doc: floating point number, api object: string or null
	Port         *uint16  `json:"port"`                     // uint16 or null
	Version      *uint    `json:"version"`                  // uint or null
	Modules      []string `json:"modules"`                  // [] len >= 0
	ExternalPort uint16   `json:"external_port"`            // uint16
	MinVer       uint     `json:"minimum_version_required"` // uint
}

// Broker defines a broker. See https://login.circonus.com/resources/api/calls/broker for more information.
type Broker struct {
	CID       string         `json:"_cid"`       // string
	Name      string         `json:"_name"`      // string
	Type      string         `json:"_type"`      // string
	Latitude  *string        `json:"_latitude"`  // string or null
	Longitude *string        `json:"_longitude"` // string or null
	Tags      []string       `json:"_tags"`      // [] len >= 0
	Details   []BrokerDetail `json:"_details"`   // [] len >= 1
}

// FetchBroker retrieves broker with passed cid.
func (a *API) FetchBroker(cid CIDType) (*Broker, error) {
	if cid == nil || *cid == "" {
		return nil, errors.Errorf("invalid broker CID (none)")
	}

	var brokerCID string
	if !strings.HasPrefix(*cid, config.BrokerPrefix) {
		brokerCID = fmt.Sprintf("%s/%s", config.BrokerPrefix, *cid)
	} else {
		brokerCID = *cid
	}

	matched, err := regexp.MatchString(config.BrokerCIDRegex, brokerCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid broker CID (%s)", brokerCID)
	}

	result, err := a.Get(brokerCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching broker")
	}

	if a.Debug {
		a.Log.Printf("fetch broker, received JSON: %s", string(result))
	}

	response := new(Broker)
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, errors.Wrap(err, "parsing broker")
	}

	return response, nil

}

// FetchBrokers returns all brokers available to the API Token.
func (a *API) FetchBrokers() (*[]Broker, error) {
	result, err := a.Get(config.BrokerPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "fetching brokers")
	}

	var response []Broker
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, errors.Wrap(err, "parsing brokers")
	}

	return &response, nil
}

// SearchBrokers returns brokers matching the specified search
// query and/or filter. If nil is passed for both parameters
// all brokers will be returned.
func (a *API) SearchBrokers(searchCriteria *SearchQueryType, filterCriteria *SearchFilterType) (*[]Broker, error) {
	q := url.Values{}

	if searchCriteria != nil && *searchCriteria != "" {
		q.Set("search", string(*searchCriteria))
	}

	if filterCriteria != nil && len(*filterCriteria) > 0 {
		for filter, criteria := range *filterCriteria {
			for _, val := range criteria {
				q.Add(filter, val)
			}
		}
	}

	if q.Encode() == "" {
		return a.FetchBrokers()
	}

	reqURL := url.URL{
		Path:     config.BrokerPrefix,
		RawQuery: q.Encode(),
	}

	result, err := a.Get(reqURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "searching brokers")
	}

	var brokers []Broker
	if err := json.Unmarshal(result, &brokers); err != nil {
		return nil, errors.Wrap(err, "parsing brokers")
	}

	return &brokers, nil
}
