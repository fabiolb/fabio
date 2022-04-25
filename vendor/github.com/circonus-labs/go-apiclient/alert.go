// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Alert API support - Fetch and Search
// See: https://login.circonus.com/resources/api/calls/alert

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

// Alert defines a alert. See https://login.circonus.com/resources/api/calls/alert for more information.
type Alert struct {
	AlertURL           string   `json:"_alert_url,omitempty"`       // string
	BrokerCID          string   `json:"_broker,omitempty"`          // string
	CheckCID           string   `json:"_check,omitempty"`           // string
	CheckName          string   `json:"_check_name,omitempty"`      // string
	CID                string   `json:"_cid,omitempty"`             // string
	Value              string   `json:"_value,omitempty"`           // string
	MetricName         string   `json:"_metric_name,omitempty"`     // string
	RuleSetCID         string   `json:"_rule_set,omitempty"`        // string
	AcknowledgementCID *string  `json:"_acknowledgement,omitempty"` // string or null
	ClearedValue       *string  `json:"_cleared_value,omitempty"`   // string or null
	MetricLinkURL      *string  `json:"_metric_link,omitempty"`     // string or null
	MetricNotes        *string  `json:"_metric_notes,omitempty"`    // string or null
	ClearedOn          *uint    `json:"_cleared_on,omitempty"`      // uint or null
	Maintenance        []string `json:"_maintenance,omitempty"`     // [] len >= 0
	Tags               []string `json:"_tags,omitempty"`            // [] len >= 0
	OccurredOn         uint     `json:"_occurred_on,omitempty"`     // uint
	Severity           uint     `json:"_severity,omitempty"`        // uint
}

// FetchAlert retrieves alert with passed cid.
func (a *API) FetchAlert(cid CIDType) (*Alert, error) {
	if cid == nil || *cid == "" {
		return nil, errors.New("invalid alert CID (none)")
	}

	var alertCID string
	if !strings.HasPrefix(*cid, config.AlertPrefix) {
		alertCID = fmt.Sprintf("%s/%s", config.AlertPrefix, *cid)
	} else {
		alertCID = *cid
	}

	matched, err := regexp.MatchString(config.AlertCIDRegex, alertCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid alert CID (%s)", alertCID)
	}

	result, err := a.Get(alertCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching alert")
	}

	if a.Debug {
		a.Log.Printf("fetch alert, received JSON: %s", string(result))
	}

	alert := &Alert{}
	if err := json.Unmarshal(result, alert); err != nil {
		return nil, errors.Wrap(err, "parsing alert")
	}

	return alert, nil
}

// FetchAlerts retrieves all alerts available to the API Token.
func (a *API) FetchAlerts() (*[]Alert, error) {
	result, err := a.Get(config.AlertPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "fetching alerts")
	}

	var alerts []Alert
	if err := json.Unmarshal(result, &alerts); err != nil {
		return nil, errors.Wrap(err, "parsing alerts")
	}

	return &alerts, nil
}

// SearchAlerts returns alerts matching the specified search query
// and/or filter. If nil is passed for both parameters all alerts
// will be returned.
func (a *API) SearchAlerts(searchCriteria *SearchQueryType, filterCriteria *SearchFilterType) (*[]Alert, error) {
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
		return a.FetchAlerts()
	}

	reqURL := url.URL{
		Path:     config.AlertPrefix,
		RawQuery: q.Encode(),
	}

	result, err := a.Get(reqURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "searching alerts")
	}

	var alerts []Alert
	if err := json.Unmarshal(result, &alerts); err != nil {
		return nil, errors.Wrap(err, "parsing alerts")
	}

	return &alerts, nil
}
