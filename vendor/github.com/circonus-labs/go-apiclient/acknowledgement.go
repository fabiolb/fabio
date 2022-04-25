// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Acknowledgement API support - Fetch, Create, Update, Delete*, and Search
// See: https://login.circonus.com/resources/api/calls/acknowledgement
// *  : delete (cancel) by updating with AcknowledgedUntil set to 0

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

// Acknowledgement defines a acknowledgement. See https://login.circonus.com/resources/api/calls/acknowledgement for more information.
type Acknowledgement struct {
	AcknowledgedUntil interface{} `json:"acknowledged_until,omitempty"` // NOTE received as uint; can be set using string or uint
	AlertCID          string      `json:"alert,omitempty"`              // string
	CID               string      `json:"_cid,omitempty"`               // string
	LastModifiedBy    string      `json:"_last_modified_by,omitempty"`  // string
	Notes             string      `json:"notes,omitempty"`              // string
	AcknowledgedBy    string      `json:"_acknowledged_by,omitempty"`   // string
	AcknowledgedOn    uint        `json:"_acknowledged_on,omitempty"`   // uint
	LastModified      uint        `json:"_last_modified,omitempty"`     // uint
	Active            bool        `json:"_active,omitempty"`            // bool
}

// NewAcknowledgement returns new Acknowledgement (with defaults, if applicable).
func NewAcknowledgement() *Acknowledgement {
	return &Acknowledgement{}
}

// FetchAcknowledgement retrieves acknowledgement with passed cid.
func (a *API) FetchAcknowledgement(cid CIDType) (*Acknowledgement, error) {
	if cid == nil || *cid == "" {
		return nil, errors.Errorf("invalid acknowledgement CID (none)")
	}

	var acknowledgementCID string
	if !strings.HasPrefix(*cid, config.AcknowledgementPrefix) {
		acknowledgementCID = fmt.Sprintf("%s/%s", config.AcknowledgementPrefix, *cid)
	} else {
		acknowledgementCID = *cid
	}

	matched, err := regexp.MatchString(config.AcknowledgementCIDRegex, acknowledgementCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid acknowledgement CID (%s)", acknowledgementCID)
	}

	result, err := a.Get(acknowledgementCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching acknowledgement")
	}

	if a.Debug {
		a.Log.Printf("fetch acknowledgement, received JSON: %s", string(result))
	}

	acknowledgement := &Acknowledgement{}
	if err := json.Unmarshal(result, acknowledgement); err != nil {
		return nil, errors.Wrap(err, "parsing acknowledgement")
	}

	return acknowledgement, nil
}

// FetchAcknowledgements retrieves all acknowledgements available to the API Token.
func (a *API) FetchAcknowledgements() (*[]Acknowledgement, error) {
	result, err := a.Get(config.AcknowledgementPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "fetching acknowledgements")
	}

	var acknowledgements []Acknowledgement
	if err := json.Unmarshal(result, &acknowledgements); err != nil {
		return nil, errors.Wrap(err, "parsing acknowledgements")
	}

	return &acknowledgements, nil
}

// UpdateAcknowledgement updates passed acknowledgement.
func (a *API) UpdateAcknowledgement(cfg *Acknowledgement) (*Acknowledgement, error) {
	if cfg == nil {
		return nil, errors.Errorf("invalid acknowledgement config (nil)")
	}

	acknowledgementCID := cfg.CID

	matched, err := regexp.MatchString(config.AcknowledgementCIDRegex, acknowledgementCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid acknowledgement CID (%s)", acknowledgementCID)
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("acknowledgement update, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Put(acknowledgementCID, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "updating acknowledgement")
	}

	acknowledgement := &Acknowledgement{}
	if err := json.Unmarshal(result, acknowledgement); err != nil {
		return nil, errors.Wrap(err, "parsing acknowledgement")
	}

	return acknowledgement, nil
}

// CreateAcknowledgement creates a new acknowledgement.
func (a *API) CreateAcknowledgement(cfg *Acknowledgement) (*Acknowledgement, error) {
	if cfg == nil {
		return nil, errors.Errorf("invalid acknowledgement config (nil)")
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	result, err := a.Post(config.AcknowledgementPrefix, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating acknowledgement")
	}

	if a.Debug {
		a.Log.Printf("acknowledgement create, sending JSON: %s", string(jsonCfg))
	}

	acknowledgement := &Acknowledgement{}
	if err := json.Unmarshal(result, acknowledgement); err != nil {
		return nil, errors.Wrap(err, "parsing acknowledgement")
	}

	return acknowledgement, nil
}

// SearchAcknowledgements returns acknowledgements matching
// the specified search query and/or filter. If nil is passed for
// both parameters all acknowledgements will be returned.
func (a *API) SearchAcknowledgements(searchCriteria *SearchQueryType, filterCriteria *SearchFilterType) (*[]Acknowledgement, error) {
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
		return a.FetchAcknowledgements()
	}

	reqURL := url.URL{
		Path:     config.AcknowledgementPrefix,
		RawQuery: q.Encode(),
	}

	result, err := a.Get(reqURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "searching acknowledgements")
	}

	var acknowledgements []Acknowledgement
	if err := json.Unmarshal(result, &acknowledgements); err != nil {
		return nil, errors.Wrap(err, "parsing acknowledgements")
	}

	return &acknowledgements, nil
}
