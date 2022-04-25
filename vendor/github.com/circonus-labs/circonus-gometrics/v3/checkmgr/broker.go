// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package checkmgr

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	apiclient "github.com/circonus-labs/go-apiclient"
	"github.com/pkg/errors"
)

// Get Broker to use when creating a check
func (cm *CheckManager) getBroker() (*apiclient.Broker, error) {
	if cm.brokerID != 0 {
		cid := fmt.Sprintf("/broker/%d", cm.brokerID)
		broker, err := cm.apih.FetchBroker(apiclient.CIDType(&cid))
		if err != nil {
			return nil, err
		}
		if !cm.isValidBroker(broker) {
			return nil, errors.Errorf(
				"error, designated broker %d [%s] is invalid (not active, does not support required check type, or connectivity issue)",
				cm.brokerID,
				broker.Name)
		}
		return broker, nil
	}
	broker, err := cm.selectBroker()
	if err != nil {
		return nil, errors.Errorf("error, unable to fetch suitable broker %s", err)
	}
	return broker, nil
}

// Get CN of Broker associated with submission_url to satisfy no IP SANS in certs
func (cm *CheckManager) getBrokerCN(broker *apiclient.Broker, submissionURL apiclient.URLType) (string, string, error) {
	u, err := url.Parse(string(submissionURL))
	if err != nil {
		return "", "", err
	}

	hostParts := strings.Split(u.Host, ":")
	host := hostParts[0]

	if net.ParseIP(host) == nil { // it's a non-ip string
		return u.Host, u.Host, nil
	}

	cn := ""
	cnList := make([]string, 0, len(broker.Details))

	for _, detail := range broker.Details {
		// broker must be active
		if detail.Status != statusActive {
			continue
		}

		// certs are generated against the CN (in theory)
		// 1. find the right broker instance with matching IP or external hostname
		// 2. set the tls.Config.ServerName to whatever that instance's CN is currently
		// 3. cert will be valid for TLS conns (in theory)
		if detail.IP != nil && *detail.IP == host {
			if cn == "" {
				cn = detail.CN
			}
			cnList = append(cnList, detail.CN)
		} else if detail.ExternalHost != nil && *detail.ExternalHost == host {
			if cn == "" {
				cn = detail.CN
			}
			cnList = append(cnList, detail.CN)
		}
	}

	if cn == "" {
		return "", "", errors.Errorf("error, unable to match URL host (%s) to Broker", u.Host)
	}

	return cn, strings.Join(cnList, ","), nil

}

// Select a broker for use when creating a check, if a specific broker
// was not specified.
func (cm *CheckManager) selectBroker() (*apiclient.Broker, error) {
	var brokerList *[]apiclient.Broker
	var err error
	enterpriseType := "enterprise"

	if len(cm.brokerSelectTag) > 0 {
		filter := apiclient.SearchFilterType{
			"f__tags_has": cm.brokerSelectTag,
		}
		brokerList, err = cm.apih.SearchBrokers(nil, &filter)
		if err != nil {
			return nil, err
		}
	} else {
		brokerList, err = cm.apih.FetchBrokers()
		if err != nil {
			return nil, err
		}
	}

	if len(*brokerList) == 0 {
		return nil, errors.New("zero brokers found")
	}

	validBrokers := make(map[string]apiclient.Broker)
	haveEnterprise := false

	for _, broker := range *brokerList {
		broker := broker
		if cm.isValidBroker(&broker) {
			validBrokers[broker.CID] = broker
			if broker.Type == enterpriseType {
				haveEnterprise = true
			}
		}
	}

	if haveEnterprise { // eliminate non-enterprise brokers from valid brokers
		for k, v := range validBrokers {
			if v.Type != enterpriseType {
				delete(validBrokers, k)
			}
		}
	}

	if len(validBrokers) == 0 {
		return nil, errors.Errorf("found %d broker(s), zero are valid", len(*brokerList))
	}

	validBrokerKeys := reflect.ValueOf(validBrokers).MapKeys()
	maxBrokers := big.NewInt(int64(len(validBrokerKeys)))
	bidx, err := rand.Int(rand.Reader, maxBrokers)
	if err != nil {
		return nil, err
	}
	selectedBroker := validBrokers[validBrokerKeys[bidx.Uint64()].String()]

	if cm.Debug {
		cm.Log.Printf("selected broker '%s'\n", selectedBroker.Name)
	}

	return &selectedBroker, nil

}

// Verify broker supports the check type to be used
func (cm *CheckManager) brokerSupportsCheckType(checkType CheckTypeType, details *apiclient.BrokerDetail) bool {

	baseType := string(checkType)

	for _, module := range details.Modules {
		if module == baseType {
			return true
		}
	}

	if idx := strings.Index(baseType, ":"); idx > 0 {
		baseType = baseType[0:idx]
	}

	for _, module := range details.Modules {
		if module == baseType {
			return true
		}
	}

	return false

}

// Is the broker valid (active, supports check type, and reachable)
func (cm *CheckManager) isValidBroker(broker *apiclient.Broker) bool {
	var brokerHost string
	var brokerPort string

	if broker.Type != "circonus" && broker.Type != "enterprise" {
		return false
	}

	valid := false

	for _, detail := range broker.Details {
		detail := detail

		// broker must be active
		if detail.Status != statusActive {
			if cm.Debug {
				cm.Log.Printf("broker '%s' is not active\n", broker.Name)
			}
			continue
		}

		// broker must have module loaded for the check type to be used
		if !cm.brokerSupportsCheckType(cm.checkType, &detail) {
			if cm.Debug {
				cm.Log.Printf("broker '%s' does not support '%s' checks\n", broker.Name, cm.checkType)
			}
			continue
		}

		if detail.ExternalPort != 0 {
			brokerPort = strconv.Itoa(int(detail.ExternalPort))
		} else {
			if detail.Port != nil && *detail.Port != 0 {
				brokerPort = strconv.Itoa(int(*detail.Port))
			} else {
				brokerPort = "43191"
			}
		}

		if detail.ExternalHost != nil && *detail.ExternalHost != "" {
			brokerHost = *detail.ExternalHost
		} else if detail.IP != nil && *detail.IP != "" {
			brokerHost = *detail.IP
		}

		if brokerHost == "" {
			cm.Log.Printf("broker '%s' instance %s has no IP or external host set", broker.Name, detail.CN)
			continue
		}

		if brokerHost == "trap.noit.circonus.net" && brokerPort != "443" {
			brokerPort = "443"
		}

		retries := 5
		for attempt := 1; attempt <= retries; attempt++ {
			// broker must be reachable and respond within designated time
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", brokerHost, brokerPort), cm.brokerMaxResponseTime)
			if err == nil {
				conn.Close()
				valid = true
				break
			}

			cm.Log.Printf("broker '%s' unable to connect, %v. Retrying in 2 seconds, attempt %d of %d\n", broker.Name, err, attempt, retries)
			time.Sleep(2 * time.Second)
		}

		if valid {
			if cm.Debug {
				cm.Log.Printf("broker '%s' is valid\n", broker.Name)
			}
			break
		}
	}
	return valid
}
