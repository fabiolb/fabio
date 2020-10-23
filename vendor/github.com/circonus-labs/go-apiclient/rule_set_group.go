// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// RuleSetGroup API support - Fetch, Create, Update, Delete, and Search
// See: https://login.circonus.com/resources/api/calls/rule_set_group

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

// RuleSetGroupFormula defines a formula for raising alerts
type RuleSetGroupFormula struct {
	Expression    interface{} `json:"expression"`     // string or uint BUG doc: string, api: string or numeric
	RaiseSeverity interface{} `json:"raise_severity"` // string or uint BUG doc: numeric, api: string or numeric
	Wait          uint        `json:"wait"`           // uint
}

// RuleSetGroupCondition defines conditions for raising alerts
type RuleSetGroupCondition struct {
	MatchingSeverities []string `json:"matching_severities"` // [] len >= 1
	RuleSetCID         string   `json:"rule_set"`            // string
}

// RuleSetGroup defines a ruleset group. See https://login.circonus.com/resources/api/calls/rule_set_group for more information.
type RuleSetGroup struct {
	CID               string                  `json:"_cid,omitempty"`      // string
	ContactGroups     map[uint8][]string      `json:"contact_groups"`      // [] len == 5
	Formulas          []RuleSetGroupFormula   `json:"formulas"`            // [] len >= 0
	Name              string                  `json:"name"`                // string
	RuleSetConditions []RuleSetGroupCondition `json:"rule_set_conditions"` // [] len >= 1
	Tags              []string                `json:"tags"`                // [] len >= 0
}

// NewRuleSetGroup returns a new RuleSetGroup (with defaults, if applicable)
func NewRuleSetGroup() *RuleSetGroup {
	return &RuleSetGroup{}
}

// FetchRuleSetGroup retrieves rule set group with passed cid.
func (a *API) FetchRuleSetGroup(cid CIDType) (*RuleSetGroup, error) {
	if cid == nil || *cid == "" {
		return nil, errors.New("invalid rule set group CID (none)")
	}

	var groupCID string
	if !strings.HasPrefix(*cid, config.RuleSetGroupPrefix) {
		groupCID = fmt.Sprintf("%s/%s", config.RuleSetGroupPrefix, *cid)
	} else {
		groupCID = *cid
	}

	matched, err := regexp.MatchString(config.RuleSetGroupCIDRegex, groupCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid rule set group CID (%s)", groupCID)
	}

	result, err := a.Get(groupCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching rule set group")
	}

	if a.Debug {
		a.Log.Printf("fetch rule set group, received JSON: %s", string(result))
	}

	rulesetGroup := &RuleSetGroup{}
	if err := json.Unmarshal(result, rulesetGroup); err != nil {
		return nil, errors.Wrap(err, "parsing rule set group")
	}

	return rulesetGroup, nil
}

// FetchRuleSetGroups retrieves all rule set groups available to API Token.
func (a *API) FetchRuleSetGroups() (*[]RuleSetGroup, error) {
	result, err := a.Get(config.RuleSetGroupPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "fetching rule set groups")
	}

	var rulesetGroups []RuleSetGroup
	if err := json.Unmarshal(result, &rulesetGroups); err != nil {
		return nil, errors.Wrap(err, "parsing rule set groups")
	}

	return &rulesetGroups, nil
}

// UpdateRuleSetGroup updates passed rule set group.
func (a *API) UpdateRuleSetGroup(cfg *RuleSetGroup) (*RuleSetGroup, error) {
	if cfg == nil {
		return nil, errors.New("invalid rule set group config (nil)")
	}

	groupCID := cfg.CID

	matched, err := regexp.MatchString(config.RuleSetGroupCIDRegex, groupCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid rule set group CID (%s)", groupCID)
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "updating rule set group")
	}

	if a.Debug {
		a.Log.Printf("update rule set group, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Put(groupCID, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "updating rule set group")
	}

	groups := &RuleSetGroup{}
	if err := json.Unmarshal(result, groups); err != nil {
		return nil, errors.Wrap(err, "parsing rule set group")
	}

	return groups, nil
}

// CreateRuleSetGroup creates a new rule set group.
func (a *API) CreateRuleSetGroup(cfg *RuleSetGroup) (*RuleSetGroup, error) {
	if cfg == nil {
		return nil, errors.New("invalid rule set group config (nil)")
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("create rule set group, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Post(config.RuleSetGroupPrefix, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "creating rule set group")
	}

	group := &RuleSetGroup{}
	if err := json.Unmarshal(result, group); err != nil {
		return nil, errors.Wrap(err, "parsing rule set group")
	}

	return group, nil
}

// DeleteRuleSetGroup deletes passed rule set group.
func (a *API) DeleteRuleSetGroup(cfg *RuleSetGroup) (bool, error) {
	if cfg == nil {
		return false, errors.New("invalid rule set group config (nil)")
	}
	return a.DeleteRuleSetGroupByCID(CIDType(&cfg.CID))
}

// DeleteRuleSetGroupByCID deletes rule set group with passed cid.
func (a *API) DeleteRuleSetGroupByCID(cid CIDType) (bool, error) {
	if cid == nil || *cid == "" {
		return false, errors.New("invalid rule set group CID (none)")
	}

	var groupCID string
	if !strings.HasPrefix(*cid, config.RuleSetGroupPrefix) {
		groupCID = fmt.Sprintf("%s/%s", config.RuleSetGroupPrefix, *cid)
	} else {
		groupCID = *cid
	}

	matched, err := regexp.MatchString(config.RuleSetGroupCIDRegex, groupCID)
	if err != nil {
		return false, err
	}
	if !matched {
		return false, errors.Errorf("invalid rule set group CID (%s)", groupCID)
	}

	_, err = a.Delete(groupCID)
	if err != nil {
		return false, errors.Wrap(err, "deleting rule set group")
	}

	return true, nil
}

// SearchRuleSetGroups returns rule set groups matching the
// specified search query and/or filter. If nil is passed for
// both parameters all rule set groups will be returned.
func (a *API) SearchRuleSetGroups(searchCriteria *SearchQueryType, filterCriteria *SearchFilterType) (*[]RuleSetGroup, error) {
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
		return a.FetchRuleSetGroups()
	}

	reqURL := url.URL{
		Path:     config.RuleSetGroupPrefix,
		RawQuery: q.Encode(),
	}

	result, err := a.Get(reqURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "searching rule set groups")
	}

	var groups []RuleSetGroup
	if err := json.Unmarshal(result, &groups); err != nil {
		return nil, errors.Wrap(err, "parsing rule set groups")
	}

	return &groups, nil
}
