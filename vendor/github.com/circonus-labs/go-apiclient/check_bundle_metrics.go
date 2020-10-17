// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// CheckBundleMetrics API support - Fetch, Create*, Update, and Delete**
// See: https://login.circonus.com/resources/api/calls/check_bundle_metrics
// *  : create metrics by adding to array with a status of 'active'
// ** : delete (distable collection of) metrics by changing status from 'active' to 'available'

package apiclient

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/circonus-labs/go-apiclient/config"
	"github.com/pkg/errors"
)

// CheckBundleMetrics defines metrics for a specific check bundle. See https://login.circonus.com/resources/api/calls/check_bundle_metrics for more information.
type CheckBundleMetrics struct {
	CID     string              `json:"_cid,omitempty"` // string
	Metrics []CheckBundleMetric `json:"metrics"`        // See check_bundle.go for CheckBundleMetric definition
}

// FetchCheckBundleMetrics retrieves metrics for the check bundle with passed cid.
func (a *API) FetchCheckBundleMetrics(cid CIDType) (*CheckBundleMetrics, error) {
	if cid == nil || *cid == "" {
		return nil, errors.New("invalid check bundle metrics CID (none)")
	}

	var metricsCID string
	if !strings.HasPrefix(*cid, config.CheckBundleMetricsPrefix) {
		metricsCID = fmt.Sprintf("%s/%s", config.CheckBundleMetricsPrefix, *cid)
	} else {
		metricsCID = *cid
	}

	matched, err := regexp.MatchString(config.CheckBundleMetricsCIDRegex, metricsCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid check bundle metrics CID (%s)", metricsCID)
	}

	result, err := a.Get(metricsCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching check bundle metrics")
	}

	if a.Debug {
		a.Log.Printf("fetch check bundle metrics, received JSON: %s", string(result))
	}

	metrics := &CheckBundleMetrics{}
	if err := json.Unmarshal(result, metrics); err != nil {
		return nil, errors.Wrap(err, "parsing check bundle metrics")
	}

	return metrics, nil
}

// UpdateCheckBundleMetrics updates passed metrics.
func (a *API) UpdateCheckBundleMetrics(cfg *CheckBundleMetrics) (*CheckBundleMetrics, error) {
	if cfg == nil {
		return nil, errors.New("invalid check bundle metrics config (nil)")
	}

	metricsCID := cfg.CID

	matched, err := regexp.MatchString(config.CheckBundleMetricsCIDRegex, metricsCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid check bundle metrics CID (%s)", metricsCID)
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("update check bundle metrics, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Put(metricsCID, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "updating check bundle metrics")
	}

	metrics := &CheckBundleMetrics{}
	if err := json.Unmarshal(result, metrics); err != nil {
		return nil, errors.Wrap(err, "parsing check bundle metrics")
	}

	return metrics, nil
}
