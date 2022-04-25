// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check bundle API support - Fetch, Create, Update, Delete, and Search
// See: https://login.circonus.com/resources/api/calls/check_bundle

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

// CheckBundleMetric individual metric configuration
type CheckBundleMetric struct {
	Name   string   `json:"name"`             // string
	Type   string   `json:"type"`             // string
	Status string   `json:"status,omitempty"` // string
	Result *string  `json:"result,omitempty"` // string or null, NOTE not settable - return/information value only
	Units  *string  `json:"units,omitempty"`  // string or null
	Tags   []string `json:"tags"`             // [] len >= 0
}

// CheckBundleConfig contains the check type specific configuration settings
// as k/v pairs (see https://login.circonus.com/resources/api/calls/check_bundle
// for the specific settings available for each distinct check type)
type CheckBundleConfig map[config.Key]string

// CheckBundle defines a check bundle. See https://login.circonus.com/resources/api/calls/check_bundle for more information.
type CheckBundle struct {
	CID                string              `json:"_cid,omitempty"`                     // string
	Status             string              `json:"status,omitempty"`                   // string
	DisplayName        string              `json:"display_name"`                       // string
	LastModifedBy      string              `json:"_last_modifed_by,omitempty"`         // string
	Target             string              `json:"target"`                             // string
	Type               string              `json:"type"`                               // string
	Notes              *string             `json:"notes,omitempty"`                    // string or null
	Config             CheckBundleConfig   `json:"config"`                             // NOTE contents of config are check type specific, map len >= 0
	Brokers            []string            `json:"brokers"`                            // [] len >= 0
	Checks             []string            `json:"_checks,omitempty"`                  // [] len >= 0
	CheckUUIDs         []string            `json:"_check_uuids,omitempty"`             // [] len >= 0
	ReverseConnectURLs []string            `json:"_reverse_connection_urls,omitempty"` // [] len >= 0
	Tags               []string            `json:"tags,omitempty"`                     // [] len >= 0
	MetricFilters      [][]string          `json:"metric_filters,omitempty"`           // [][type,rule_regx,comment]
	Metrics            []CheckBundleMetric `json:"metrics"`                            // [] >= 0
	Timeout            float32             `json:"timeout,omitempty"`                  // float32
	Period             uint                `json:"period,omitempty"`                   // uint
	Created            uint                `json:"_created,omitempty"`                 // uint
	LastModified       uint                `json:"_last_modified,omitempty"`           // uint
	MetricLimit        int                 `json:"metric_limit,omitempty"`             // int
}

// NewCheckBundle returns new CheckBundle (with defaults, if applicable)
func NewCheckBundle() *CheckBundle {
	return &CheckBundle{
		Config:      make(CheckBundleConfig, config.DefaultConfigOptionsSize),
		MetricLimit: config.DefaultCheckBundleMetricLimit,
		Period:      config.DefaultCheckBundlePeriod,
		Timeout:     config.DefaultCheckBundleTimeout,
		Status:      config.DefaultCheckBundleStatus,
	}
}

// FetchCheckBundle retrieves check bundle with passed cid.
func (a *API) FetchCheckBundle(cid CIDType) (*CheckBundle, error) {
	if cid == nil || *cid == "" {
		return nil, errors.New("invalid check bundle CID (none)")
	}

	var bundleCID string
	if !strings.HasPrefix(*cid, config.CheckBundlePrefix) {
		bundleCID = fmt.Sprintf("%s/%s", config.CheckBundlePrefix, *cid)
	} else {
		bundleCID = *cid
	}

	matched, err := regexp.MatchString(config.CheckBundleCIDRegex, bundleCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid check bundle CID (%v)", bundleCID)
	}

	result, err := a.Get(bundleCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching check bundle")
	}

	if a.Debug {
		a.Log.Printf("fetch check bundle, received JSON: %s", string(result))
	}

	checkBundle := &CheckBundle{}
	if err := json.Unmarshal(result, checkBundle); err != nil {
		return nil, errors.Wrap(err, "parsing check bundle")
	}

	return checkBundle, nil
}

// FetchCheckBundles retrieves all check bundles available to the API Token.
func (a *API) FetchCheckBundles() (*[]CheckBundle, error) {
	result, err := a.Get(config.CheckBundlePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "fetching check bundles")
	}

	var checkBundles []CheckBundle
	if err := json.Unmarshal(result, &checkBundles); err != nil {
		return nil, errors.Wrap(err, "parsing check bundles")
	}

	return &checkBundles, nil
}

// UpdateCheckBundle updates passed check bundle.
func (a *API) UpdateCheckBundle(cfg *CheckBundle) (*CheckBundle, error) {
	if cfg == nil {
		return nil, errors.New("invalid check bundle config (nil)")
	}

	bundleCID := cfg.CID

	matched, err := regexp.MatchString(config.CheckBundleCIDRegex, bundleCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid check bundle CID (%s)", bundleCID)
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("update check bundle, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Put(bundleCID, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "updating check bundle")
	}

	checkBundle := &CheckBundle{}
	if err := json.Unmarshal(result, checkBundle); err != nil {
		return nil, errors.Wrap(err, "parsing check bundle")
	}

	return checkBundle, nil
}

// CreateCheckBundle creates a new check bundle (check).
func (a *API) CreateCheckBundle(cfg *CheckBundle) (*CheckBundle, error) {
	if cfg == nil {
		return nil, errors.New("invalid check bundle config (nil)")
	}

	if len(cfg.Tags) > 0 {
		// remove blanks
		tags := make([]string, 0, len(cfg.Tags))
		for _, tag := range cfg.Tags {
			if tag != "" {
				tags = append(tags, tag)
			}
		}
		cfg.Tags = tags
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("create check bundle, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Post(config.CheckBundlePrefix, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating check bundle")
	}

	checkBundle := &CheckBundle{}
	if err := json.Unmarshal(result, checkBundle); err != nil {
		return nil, errors.Wrap(err, "parsing check bundle")
	}

	return checkBundle, nil
}

// DeleteCheckBundle deletes passed check bundle.
func (a *API) DeleteCheckBundle(cfg *CheckBundle) (bool, error) {
	if cfg == nil {
		return false, errors.New("invalid check bundle config (nil)")
	}
	return a.DeleteCheckBundleByCID(CIDType(&cfg.CID))
}

// DeleteCheckBundleByCID deletes check bundle with passed cid.
func (a *API) DeleteCheckBundleByCID(cid CIDType) (bool, error) {

	if cid == nil || *cid == "" {
		return false, errors.New("invalid check bundle CID (none)")
	}

	var bundleCID string
	if !strings.HasPrefix(*cid, config.CheckBundlePrefix) {
		bundleCID = fmt.Sprintf("%s/%s", config.CheckBundlePrefix, *cid)
	} else {
		bundleCID = *cid
	}

	matched, err := regexp.MatchString(config.CheckBundleCIDRegex, bundleCID)
	if err != nil {
		return false, err
	}
	if !matched {
		return false, errors.Errorf("invalid check bundle CID (%v)", bundleCID)
	}

	_, err = a.Delete(bundleCID)
	if err != nil {
		return false, errors.Wrap(err, "deleting check bundle")
	}

	return true, nil
}

// SearchCheckBundles returns check bundles matching the specified
// search query and/or filter. If nil is passed for both parameters
// all check bundles will be returned.
func (a *API) SearchCheckBundles(searchCriteria *SearchQueryType, filterCriteria *SearchFilterType) (*[]CheckBundle, error) {

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
		return a.FetchCheckBundles()
	}

	reqURL := url.URL{
		Path:     config.CheckBundlePrefix,
		RawQuery: q.Encode(),
	}

	resp, err := a.Get(reqURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "searching check bundles")
	}

	var results []CheckBundle
	if err := json.Unmarshal(resp, &results); err != nil {
		return nil, errors.Wrap(err, "parsing check bundles")
	}

	return &results, nil
}
