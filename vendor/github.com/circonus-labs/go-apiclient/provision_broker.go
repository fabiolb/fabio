// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ProvisionBroker API support - Fetch, Create, and Update
// See: https://login.circonus.com/resources/api/calls/provision_broker
// Note that the provision_broker endpoint does not return standard cid format
//      of '/object/item' (e.g. /provision_broker/abc-123) it just returns 'item'

package apiclient

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/circonus-labs/go-apiclient/config"
	"github.com/pkg/errors"
)

// BrokerStratcon defines stratcons for broker
type BrokerStratcon struct {
	CN   string `json:"cn,omitempty"`   // string
	Host string `json:"host,omitempty"` // string
	Port string `json:"port,omitempty"` // string
}

// ProvisionBroker defines a provision broker [request]. See https://login.circonus.com/resources/api/calls/provision_broker for more details.
type ProvisionBroker struct {
	Cert                    string           `json:"_cert,omitempty"`                     // string
	CID                     string           `json:"_cid,omitempty"`                      // string
	CSR                     string           `json:"_csr,omitempty"`                      // string
	ExternalHost            string           `json:"external_host,omitempty"`             // string
	ExternalPort            string           `json:"external_port,omitempty"`             // string
	IPAddress               string           `json:"ipaddress,omitempty"`                 // string
	Latitude                string           `json:"latitude,omitempty"`                  // string
	Longitude               string           `json:"longitude,omitempty"`                 // string
	Name                    string           `json:"noit_name,omitempty"`                 // string
	Port                    string           `json:"port,omitempty"`                      // string
	PreferReverseConnection bool             `json:"prefer_reverse_connection,omitempty"` // boolean
	Rebuild                 bool             `json:"rebuild,omitempty"`                   // boolean
	Stratcons               []BrokerStratcon `json:"_stratcons,omitempty"`                // [] len >= 1
	Tags                    []string         `json:"tags,omitempty"`                      // [] len >= 0
}

// NewProvisionBroker returns a new ProvisionBroker (with defaults, if applicable)
func NewProvisionBroker() *ProvisionBroker {
	return &ProvisionBroker{}
}

// FetchProvisionBroker retrieves provision broker [request] with passed cid.
func (a *API) FetchProvisionBroker(cid CIDType) (*ProvisionBroker, error) {
	if cid == nil || *cid == "" {
		return nil, errors.New("invalid provision broker CID (none)")
	}

	var brokerCID string
	if !strings.HasPrefix(*cid, config.ProvisionBrokerPrefix) {
		brokerCID = fmt.Sprintf("%s/%s", config.ProvisionBrokerPrefix, *cid)
	} else {
		brokerCID = *cid
	}

	matched, err := regexp.MatchString(config.ProvisionBrokerCIDRegex, brokerCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid provision broker CID (%s)", brokerCID)
	}

	result, err := a.Get(brokerCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching provision broker")
	}

	if a.Debug {
		a.Log.Printf("fetch broker provision request, received JSON: %s", string(result))
	}

	broker := &ProvisionBroker{}
	if err := json.Unmarshal(result, broker); err != nil {
		return nil, errors.Wrap(err, "parsing provision broker")
	}

	return broker, nil
}

// UpdateProvisionBroker updates a broker definition [request].
func (a *API) UpdateProvisionBroker(cid CIDType, cfg *ProvisionBroker) (*ProvisionBroker, error) {
	if cid == nil || *cid == "" {
		return nil, errors.New("invalid provision broker CID (none)")
	}

	if cfg == nil {
		return nil, errors.New("invalid provision broker config (nil)")
	}

	brokerCID := *cid

	matched, err := regexp.MatchString(config.ProvisionBrokerCIDRegex, brokerCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid provision broker CID (%s)", brokerCID)
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("update broker provision request, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Put(brokerCID, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "updating provision broker")
	}

	broker := &ProvisionBroker{}
	if err := json.Unmarshal(result, broker); err != nil {
		return nil, errors.Wrap(err, "parsing provision broker")
	}

	return broker, nil
}

// CreateProvisionBroker creates a new provison broker [request].
func (a *API) CreateProvisionBroker(cfg *ProvisionBroker) (*ProvisionBroker, error) {
	if cfg == nil {
		return nil, errors.New("invalid provision broker config (nil)")
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("create broker provision request, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Post(config.ProvisionBrokerPrefix, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating provision broker")
	}

	broker := &ProvisionBroker{}
	if err := json.Unmarshal(result, broker); err != nil {
		return nil, errors.Wrap(err, "parsing provision broker")
	}

	return broker, nil
}
