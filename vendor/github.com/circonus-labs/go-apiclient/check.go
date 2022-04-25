// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check API support - Fetch and Search
// See: https://login.circonus.com/resources/api/calls/check
// Notes: checks do not directly support create, update, and delete - see check bundle.

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

// CheckDetails contains [undocumented] check type specific information
type CheckDetails map[config.Key]string

// Check defines a check. See https://login.circonus.com/resources/api/calls/check for more information.
type Check struct {
	BrokerCID      string       `json:"_broker"`       // string
	CheckBundleCID string       `json:"_check_bundle"` // string
	CheckUUID      string       `json:"_check_uuid"`   // string
	CID            string       `json:"_cid"`          // string
	Details        CheckDetails `json:"_details"`      // NOTE contents of details are check type specific, map len >= 0
	ReverseURLs    []string     `json:"_reverse_urls"` // []string list of reverse urls (one per broker in cluster)
	Active         bool         `json:"_active"`       // bool
}

// FetchCheck retrieves check with passed cid.
func (a *API) FetchCheck(cid CIDType) (*Check, error) {
	if cid == nil || *cid == "" {
		return nil, errors.New("invalid check CID (none)")
	}

	var checkCID string
	if !strings.HasPrefix(*cid, config.CheckPrefix) {
		checkCID = fmt.Sprintf("%s/%s", config.CheckPrefix, *cid)
	} else {
		checkCID = *cid
	}

	matched, err := regexp.MatchString(config.CheckCIDRegex, checkCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid check CID (%s)", checkCID)
	}

	result, err := a.Get(checkCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching check")
	}

	if a.Debug {
		a.Log.Printf("fetch check, received JSON: %s", string(result))
	}

	check := new(Check)
	if err := json.Unmarshal(result, check); err != nil {
		return nil, errors.Wrap(err, "parsing check")
	}

	return check, nil
}

// FetchChecks retrieves all checks available to the API Token.
func (a *API) FetchChecks() (*[]Check, error) {
	result, err := a.Get(config.CheckPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "fetching checks")
	}

	var checks []Check
	if err := json.Unmarshal(result, &checks); err != nil {
		return nil, errors.Wrap(err, "parsing checks")
	}

	return &checks, nil
}

// SearchChecks returns checks matching the specified search query
// and/or filter. If nil is passed for both parameters all checks
// will be returned.
func (a *API) SearchChecks(searchCriteria *SearchQueryType, filterCriteria *SearchFilterType) (*[]Check, error) {
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
		return a.FetchChecks()
	}

	reqURL := url.URL{
		Path:     config.CheckPrefix,
		RawQuery: q.Encode(),
	}

	result, err := a.Get(reqURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "searching checks")
	}

	var checks []Check
	if err := json.Unmarshal(result, &checks); err != nil {
		return nil, errors.Wrap(err, "parsing checks")
	}

	return &checks, nil
}
