// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package checkmgr

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/circonus-labs/go-apiclient"
	"github.com/circonus-labs/go-apiclient/config"
	"github.com/pkg/errors"
)

// UpdateCheck determines if the check needs to be updated (new metrics, tags, etc.)
func (cm *CheckManager) UpdateCheck(newMetrics map[string]*apiclient.CheckBundleMetric) {
	// only if check manager is enabled
	if !cm.enabled {
		return
	}

	// only if checkBundle has been populated
	if cm.checkBundle == nil {
		return
	}

	// only if there is *something* to update
	if !cm.forceCheckUpdate && len(newMetrics) == 0 && len(cm.metricTags) == 0 {
		return
	}

	// refresh check bundle (in case there were changes made by other apps or in UI)
	cid := cm.checkBundle.CID
	checkBundle, err := cm.apih.FetchCheckBundle(apiclient.CIDType(&cid))
	if err != nil {
		cm.Log.Printf("error fetching up-to-date check bundle %v", err)
		return
	}
	cm.cbmu.Lock()
	cm.checkBundle = checkBundle
	cm.cbmu.Unlock()

	// check metric_limit and see if it’s 0, if so, don't even bother to try to update the check.

	cm.addNewMetrics(newMetrics)

	if len(cm.metricTags) > 0 {
		// note: if a tag has been added (queued) for a metric which never gets sent
		//       the tags will be discarded. (setting tags does not *create* metrics.)
		for metricName, metricTags := range cm.metricTags {
			for metricIdx, metric := range cm.checkBundle.Metrics {
				if metric.Name == metricName {
					cm.checkBundle.Metrics[metricIdx].Tags = metricTags
					break
				}
			}
			cm.mtmu.Lock()
			delete(cm.metricTags, metricName)
			cm.mtmu.Unlock()
		}
		cm.forceCheckUpdate = true
	}

	if cm.forceCheckUpdate {
		newCheckBundle, err := cm.apih.UpdateCheckBundle(cm.checkBundle)
		if err != nil {
			cm.Log.Printf("error updating check bundle %v", err)
			return
		}

		cm.forceCheckUpdate = false
		cm.cbmu.Lock()
		cm.checkBundle = newCheckBundle
		cm.cbmu.Unlock()
		cm.inventoryMetrics()
	}

}

// Initialize CirconusMetrics instance. Attempt to find a check otherwise create one.
// use cases:
//
// check [bundle] by submission url
// check [bundle] by *check* id (note, not check_bundle id)
// check [bundle] by search
// create check [bundle]
func (cm *CheckManager) initializeTrapURL() error {
	if cm.trapURL != "" {
		return nil
	}

	cm.trapmu.Lock()
	defer cm.trapmu.Unlock()

	// special case short-circuit: just send to a url, no check management
	// up to user to ensure that if url is https that it will work (e.g. not self-signed)
	if cm.checkSubmissionURL != "" {
		if !cm.enabled {
			cm.trapURL = cm.checkSubmissionURL
			cm.trapLastUpdate = time.Now()
			return nil
		}
	}

	if !cm.enabled {
		return errors.New("unable to initialize trap, check manager is disabled")
	}

	var err error
	var check *apiclient.Check
	var checkBundle *apiclient.CheckBundle
	var broker *apiclient.Broker

	switch {
	case cm.checkSubmissionURL != "":
		check, err = cm.fetchCheckBySubmissionURL(cm.checkSubmissionURL)
		if err != nil {
			return err
		}
		if !check.Active {
			return errors.Errorf("error, check %v is not active", check.CID)
		}
		// extract check id from check object returned from looking up using submission url
		// set m.CheckId to the id
		// set m.SubmissionUrl to "" to prevent trying to search on it going forward
		// use case: if the broker is changed in the UI metrics would stop flowing
		// unless the new submission url can be fetched with the API (which is no
		// longer possible using the original submission url)
		var id int
		id, err = strconv.Atoi(strings.ReplaceAll(check.CID, "/check/", ""))
		if err == nil {
			cm.checkID = apiclient.IDType(id)
			cm.checkSubmissionURL = ""
		} else {
			cm.Log.Printf("SubmissionUrl check CID to Check ID: unable to convert %s to int %q\n", check.CID, err)
		}
	case cm.checkID > 0:
		cid := fmt.Sprintf("/check/%d", cm.checkID)
		check, err = cm.apih.FetchCheck(apiclient.CIDType(&cid))
		if err != nil {
			return err
		}
		if !check.Active {
			return errors.Errorf("error, check %v is not active", check.CID)
		}
	default:
		// new search (check.target != instanceid, instanceid encoded in notes field)
		searchCriteria := fmt.Sprintf(
			"(active:1)(type:\"%s\")(tags:%s)", cm.checkType, strings.Join(cm.checkSearchTag, ","))
		filterCriteria := map[string][]string{"f_notes": {*cm.getNotes()}}
		checkBundle, err = cm.checkBundleSearch(searchCriteria, filterCriteria)
		if err != nil {
			return err
		}

		if checkBundle == nil {
			// old search (instanceid as check.target)
			searchCriteria := fmt.Sprintf(
				"(active:1)(type:\"%s\")(host:\"%s\")(tags:%s)", cm.checkType, cm.checkTarget, strings.Join(cm.checkSearchTag, ","))
			checkBundle, err = cm.checkBundleSearch(searchCriteria, map[string][]string{})
			if err != nil {
				return err
			}
		}

		if checkBundle == nil {
			// err==nil && checkBundle==nil is "no check bundles matched"
			// an error *should* be returned for any other invalid scenario
			checkBundle, broker, err = cm.createNewCheck()
			if err != nil {
				return err
			}
		}
	}

	if checkBundle == nil {
		if check != nil {
			cid := check.CheckBundleCID
			checkBundle, err = cm.apih.FetchCheckBundle(apiclient.CIDType(&cid))
			if err != nil {
				return err
			}
		} else {
			return errors.Errorf("error, unable to retrieve, find, or create a check bundle")
		}
	}

	if broker == nil {
		cid := checkBundle.Brokers[0]
		broker, err = cm.apih.FetchBroker(apiclient.CIDType(&cid))
		if err != nil {
			return err
		}
	}

	// retain to facilitate metric management (adding new metrics specifically)
	cm.checkBundle = checkBundle

	// determine the trap url to which metrics should be PUT
	if strings.HasPrefix(checkBundle.Type, "httptrap") {
		if turl, found := checkBundle.Config[config.SubmissionURL]; found {
			cm.trapURL = apiclient.URLType(turl)
		} else {
			if cm.Debug {
				cm.Log.Printf("missing config.%s %+v", config.SubmissionURL, checkBundle)
			}
			return errors.Errorf("error, unable to use check, no %s in config", config.SubmissionURL)
		}
	} else {
		// build a submission_url for non-httptrap checks out of mtev_reverse url
		if len(checkBundle.ReverseConnectURLs) == 0 {
			return errors.Errorf("error, %s is not an HTTPTRAP check and no reverse connection urls found", checkBundle.Checks[0])
		}
		mtevURL := checkBundle.ReverseConnectURLs[0]
		mtevURL = strings.Replace(mtevURL, "mtev_reverse", "https", 1)
		mtevURL = strings.Replace(mtevURL, "check", "module/httptrap", 1)
		if rs, found := checkBundle.Config[config.ReverseSecretKey]; found {
			cm.trapURL = apiclient.URLType(fmt.Sprintf("%s/%s", mtevURL, rs))
		} else {
			if cm.Debug {
				cm.Log.Printf("missing config.%s %+v", config.ReverseSecretKey, checkBundle)
			}
			return errors.Errorf("error, unable to use check, no %s in config", config.ReverseSecretKey)
		}
	}

	// used when sending as "ServerName" get around certs not having IP SANS
	// (cert created with server name as CN but IP used in trap url)
	cn, cnList, err := cm.getBrokerCN(broker, cm.trapURL)
	if err != nil {
		return err
	}
	cm.trapCN = BrokerCNType(cn)
	cm.trapCNList = cnList

	if cm.enabled {
		u, err := url.Parse(string(cm.trapURL))
		if err != nil {
			return err
		}
		if u.Scheme == "https" {
			if err := cm.loadCACert(); err != nil {
				return err
			}
		}
	}

	// check is using metric filters, disable check management
	cm.manageMetrics = true
	if len(cm.checkBundle.MetricFilters) > 0 {
		cm.manageMetrics = false
	}
	if cm.manageMetrics {
		cm.inventoryMetrics()
	}
	cm.trapLastUpdate = time.Now()

	return nil
}

// Search for a check bundle given a predetermined set of criteria
func (cm *CheckManager) checkBundleSearch(searchCriteria string, filterCriteria map[string][]string) (*apiclient.CheckBundle, error) {
	search := apiclient.SearchQueryType(searchCriteria)
	filter := apiclient.SearchFilterType(filterCriteria)
	checkBundles, err := cm.apih.SearchCheckBundles(&search, &filter)
	if err != nil {
		return nil, err
	}

	if len(*checkBundles) == 0 {
		return nil, nil // trigger creation of a new check
	}

	numActive := 0
	checkID := -1

	for idx, check := range *checkBundles {
		if check.Status == statusActive {
			numActive++
			checkID = idx
		}
	}

	if numActive > 1 {
		return nil, errors.Errorf("multiple check bundles match criteria - search(%v) filter(%v)", searchCriteria, filterCriteria)
	}

	bundle := (*checkBundles)[checkID]

	return &bundle, nil
}

// Create a new check to receive metrics
func (cm *CheckManager) createNewCheck() (*apiclient.CheckBundle, *apiclient.Broker, error) {
	checkSecret := string(cm.checkSecret)
	if checkSecret == "" {
		secret, err := cm.makeSecret()
		if err != nil {
			secret = "myS3cr3t"
		}
		checkSecret = secret
	}

	broker, err := cm.getBroker()
	if err != nil {
		return nil, nil, err
	}

	chkcfg := &apiclient.CheckBundle{
		Brokers:       []string{broker.CID},
		Config:        make(map[config.Key]string),
		DisplayName:   string(cm.checkDisplayName),
		MetricFilters: [][]string{{"deny", "^$", ""}, {"allow", "^.+$", ""}},
		MetricLimit:   config.DefaultCheckBundleMetricLimit,
		Metrics:       []apiclient.CheckBundleMetric{},
		Notes:         cm.getNotes(),
		Period:        60,
		Status:        statusActive,
		Tags:          append(cm.checkSearchTag, cm.checkTags...),
		Target:        string(cm.checkTarget),
		Timeout:       10,
		Type:          string(cm.checkType),
	}

	if len(cm.customConfigFields) > 0 {
		for fld, val := range cm.customConfigFields {
			chkcfg.Config[config.Key(fld)] = val
		}
	}

	//
	// use the default config settings if these are NOT set by user configuration
	//
	if val, ok := chkcfg.Config[config.AsyncMetrics]; !ok || val == "" {
		chkcfg.Config[config.AsyncMetrics] = "true"
	}

	if val, ok := chkcfg.Config[config.Secret]; !ok || val == "" {
		chkcfg.Config[config.Secret] = checkSecret
	}

	// set metric filters if provided
	if len(cm.checkMetricFilters) > 0 {
		mf := make([][]string, len(cm.checkMetricFilters))
		for idx, rule := range cm.checkMetricFilters {
			mf[idx] = []string{rule.Type, rule.Filter, rule.Comment}
		}
		chkcfg.MetricFilters = mf
	}

	checkBundle, err := cm.apih.CreateCheckBundle(chkcfg)
	if err != nil {
		return nil, nil, err
	}

	return checkBundle, broker, nil
}

// Create a dynamic secret to use with a new check
func (cm *CheckManager) makeSecret() (string, error) {
	hash := sha256.New()
	x := make([]byte, 2048)
	if _, err := rand.Read(x); err != nil {
		return "", err
	}
	if _, err := hash.Write(x); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil))[0:16], nil
}

func (cm *CheckManager) getNotes() *string {
	notes := fmt.Sprintf("cgm_instanceid|%s", cm.checkInstanceID)
	return &notes
}

// FetchCheckBySubmissionURL fetch a check configuration by submission_url
func (cm *CheckManager) fetchCheckBySubmissionURL(submissionURL apiclient.URLType) (*apiclient.Check, error) {
	if string(submissionURL) == "" {
		return nil, errors.New("error, invalid submission URL (blank)")
	}

	u, err := url.Parse(string(submissionURL))
	if err != nil {
		return nil, err
	}

	// valid trap url: scheme://host[:port]/module/httptrap/UUID/secret

	// does it smell like a valid trap url path
	if !strings.Contains(u.Path, "/module/httptrap/") {
		return nil, errors.Errorf("error, invalid submission URL '%s', unrecognized path", submissionURL)
	}

	// extract uuid
	pathParts := strings.Split(strings.Replace(u.Path, "/module/httptrap/", "", 1), "/")
	if len(pathParts) != 2 {
		return nil, errors.Errorf("error, invalid submission URL '%s', UUID not where expected", submissionURL)
	}
	uuid := pathParts[0]

	filter := apiclient.SearchFilterType{"f__check_uuid": []string{uuid}}

	checks, err := cm.apih.SearchChecks(nil, &filter)
	if err != nil {
		return nil, err
	}

	if len(*checks) == 0 {
		return nil, errors.Errorf("error, no checks found with UUID %s", uuid)
	}

	numActive := 0
	checkID := -1

	for idx, check := range *checks {
		if check.Active {
			numActive++
			if checkID == -1 {
				checkID = idx
			}
		}
	}

	if checkID == -1 {
		return nil, errors.Errorf("error, no active checks found %v", *checks)
	}

	if numActive > 1 {
		return nil, errors.Errorf("error, multiple checks with same UUID %s", uuid)
	}

	check := (*checks)[checkID]

	return &check, nil
}
