// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Account API support - Fetch and Update
// See: https://login.circonus.com/resources/api/calls/account
// Note: Create and Delete are not supported for Accounts via the API

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

// AccountLimit defines a usage limit imposed on account
type AccountLimit struct {
	Type  string `json:"_type,omitempty"`  // string
	Limit uint   `json:"_limit,omitempty"` // uint >=0
	Used  uint   `json:"_used,omitempty"`  // uint >=0
}

// AccountInvite defines outstanding invites
type AccountInvite struct {
	Email string `json:"email"` // string
	Role  string `json:"role"`  // string
}

// AccountUser defines current users
type AccountUser struct {
	Role    string `json:"role"` // string
	UserCID string `json:"user"` // string
}

// Account defines an account. See https://login.circonus.com/resources/api/calls/account for more information.
type Account struct {
	CID           string          `json:"_cid,omitempty"`            // string
	Country       string          `json:"country_code,omitempty"`    // string
	Name          string          `json:"name,omitempty"`            // string
	OwnerCID      string          `json:"_owner,omitempty"`          // string
	Timezone      string          `json:"timezone,omitempty"`        // string
	UIBaseURL     string          `json:"_ui_base_url,omitempty"`    // string
	Address1      *string         `json:"address1,omitempty"`        // string or null
	Address2      *string         `json:"address2,omitempty"`        // string or null
	CCEmail       *string         `json:"cc_email,omitempty"`        // string or null
	City          *string         `json:"city,omitempty"`            // string or null
	Description   *string         `json:"description,omitempty"`     // string or null
	StateProv     *string         `json:"state_prov,omitempty"`      // string or null
	ContactGroups []string        `json:"_contact_groups,omitempty"` // [] len >= 0
	Invites       []AccountInvite `json:"invites,omitempty"`         // [] len >= 0
	Usage         []AccountLimit  `json:"_usage,omitempty"`          // [] len >= 0
	Users         []AccountUser   `json:"users,omitempty"`           // [] len >= 0
}

// FetchAccount retrieves account with passed cid. Pass nil for '/account/current'.
func (a *API) FetchAccount(cid CIDType) (*Account, error) {
	var accountCID string

	switch {
	case cid == nil:
		fallthrough
	case *cid == "":
		accountCID = config.AccountPrefix + "/current"
	case !strings.HasPrefix(*cid, config.AccountPrefix):
		accountCID = fmt.Sprintf("%s/%s", config.AccountPrefix, *cid)
	default:
		accountCID = *cid
	}

	matched, err := regexp.MatchString(config.AccountCIDRegex, accountCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid account CID (%s)", accountCID)
	}

	result, err := a.Get(accountCID)
	if err != nil {
		return nil, errors.Wrap(err, "fetching account")
	}

	if a.Debug {
		a.Log.Printf("fetch account, received JSON: %s", string(result))
	}

	account := new(Account)
	if err := json.Unmarshal(result, account); err != nil {
		return nil, errors.Wrap(err, "parsing account")
	}

	return account, nil
}

// FetchAccounts retrieves all accounts available to the API Token.
func (a *API) FetchAccounts() (*[]Account, error) {
	result, err := a.Get(config.AccountPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "fetching accounts")
	}

	var accounts []Account
	if err := json.Unmarshal(result, &accounts); err != nil {
		return nil, errors.Wrap(err, "parsing accounts")
	}

	return &accounts, nil
}

// UpdateAccount updates passed account.
func (a *API) UpdateAccount(cfg *Account) (*Account, error) {
	if cfg == nil {
		return nil, errors.Errorf("invalid account config (nil)")
	}

	accountCID := cfg.CID

	matched, err := regexp.MatchString(config.AccountCIDRegex, accountCID)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, errors.Errorf("invalid account CID (%s)", accountCID)
	}

	jsonCfg, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if a.Debug {
		a.Log.Printf("account update, sending JSON: %s", string(jsonCfg))
	}

	result, err := a.Put(accountCID, jsonCfg)
	if err != nil {
		return nil, errors.Wrap(err, "updating account")
	}

	account := &Account{}
	if err := json.Unmarshal(result, account); err != nil {
		return nil, errors.Wrap(err, "parsing account")
	}

	return account, nil
}

// SearchAccounts returns accounts matching a filter (search queries are not
// supported by the account endpoint). Pass nil as filter for all accounts the
// API Token can access.
func (a *API) SearchAccounts(filterCriteria *SearchFilterType) (*[]Account, error) {
	q := url.Values{}

	if filterCriteria != nil && len(*filterCriteria) > 0 {
		for filter, criteria := range *filterCriteria {
			for _, val := range criteria {
				q.Add(filter, val)
			}
		}
	}

	if q.Encode() == "" {
		return a.FetchAccounts()
	}

	reqURL := url.URL{
		Path:     config.AccountPrefix,
		RawQuery: q.Encode(),
	}

	result, err := a.Get(reqURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "searching accounts")
	}

	var accounts []Account
	if err := json.Unmarshal(result, &accounts); err != nil {
		return nil, err
	}

	return &accounts, nil
}
