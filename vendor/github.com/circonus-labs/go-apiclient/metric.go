// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Metric API support - Fetch, Create*, Update, Delete*, and Search
// See: https://login.circonus.com/resources/api/calls/metric
// *  : create and delete are handled via check_bundle or check_bundle_metrics

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

// Metric defines a metric. See https://login.circonus.com/resources/api/calls/metric for more information.
type Metric struct {
	Notes          *string  `json:"notes,omitempty"`         // string or null
	Link           *string  `json:"link,omitempty"`          // string or null
	CheckBundleCID string   `json:"_check_bundle,omitempty"` // string
	CheckCID       string   `json:"_check,omitempty"`        // string
	CheckUUID      string   `json:"_check_uuid,omitempty"`   // string
	CID            string   `json:"_cid,omitempty"`          // string
	Histogram      string   `json:"_histogram,omitempty"`    // string
	MetricName     string   `json:"_metric_name,omitempty"`  // string
	MetricType     string   `json:"_metric_type,omitempty"`  // string
	CheckTags      []string `json:"_check_tags,omitempty"`   // [] len >= 0
	Active         bool     `json:"_active,omitempty"`       // boolean
	CheckActive    bool     `json:"_check_active,omitempty"` // boolean
	// DEPRECATED
	// Tags           []string `json:"tags,omitempty"`          // [] len >= 0
	// Units          *string  `json:"units,omitempty"`         // string or null
}

// FetchMetric retrieves metric with passed cid.
func (a *API) FetchMetric(cid CIDType) (*Metric, error) {
	if cid == nil || *cid == "" {
		return nil, errors.New("invalid metric CID (none)")
	}

	var metricCID string
	if !strings.HasPrefix(*cid, config.MetricPrefix) {
		metricCID = fmt.Sprintf("%s/%s", config.MetricPrefix, *cid)
	} else {
		metricCID = *cid
	}

	matched, err := regexp.MatchString(config.MetricCIDRegex, metricCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid metric CID (%s)", metricCID)
	}

	result, err := a.Get(metricCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching metric")
	}

	if a.Debug {
		a.Log.Printf("fetch metric, received JSON: %s", string(result))
	}

	metric := &Metric{}
	if err := json.Unmarshal(result, metric); err != nil {
		return nil, errors.Wrap(err, "parsing metric")
	}

	return metric, nil
}

// FetchMetrics retrieves all metrics available to API Token.
func (a *API) FetchMetrics() (*[]Metric, error) {
	result, err := a.Get(config.MetricPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "fetching metrics")
	}

	var metrics []Metric
	if err := json.Unmarshal(result, &metrics); err != nil {
		return nil, errors.Wrap(err, "parsing metrics")
	}

	return &metrics, nil
}

// UpdateMetric updates passed metric.
func (a *API) UpdateMetric(cfg *Metric) (*Metric, error) {
	if cfg == nil {
		return nil, errors.New("invalid metric config (nil)")
	}

	metricCID := cfg.CID

	matched, err := regexp.MatchString(config.MetricCIDRegex, metricCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid metric CID (%s)", metricCID)
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("update metric, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Put(metricCID, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "updating metric")
	}

	metric := &Metric{}
	if err := json.Unmarshal(result, metric); err != nil {
		return nil, errors.Wrap(err, "parsing metric")
	}

	return metric, nil
}

// SearchMetrics returns metrics matching the specified search query
// and/or filter. If nil is passed for both parameters all metrics
// will be returned.
func (a *API) SearchMetrics(searchCriteria *SearchQueryType, filterCriteria *SearchFilterType) (*[]Metric, error) {
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
		return a.FetchMetrics()
	}

	reqURL := url.URL{
		Path:     config.MetricPrefix,
		RawQuery: q.Encode(),
	}

	result, err := a.Get(reqURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "searching metrics")
	}

	var metrics []Metric
	if err := json.Unmarshal(result, &metrics); err != nil {
		return nil, errors.Wrap(err, "parsing metrics")
	}

	return &metrics, nil
}
