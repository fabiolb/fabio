// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package circonusgometrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/circonus-labs/go-apiclient"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
)

type trapResult struct {
	Error    string        `json:"error,omitempty"`
	Duration time.Duration // doesn't come from broker
	Filtered uint64        `json:"filtered,omitempty"`
	Stats    uint64        `json:"stats"`
}

func (tr *trapResult) String() string {
	ret := fmt.Sprintf("stats: %d, filtered: %d", tr.Stats, tr.Filtered)
	if tr.Error != "" {
		ret += ", error: " + tr.Error
	}
	return ret
}
func (m *CirconusMetrics) submit(output Metrics, newMetrics map[string]*apiclient.CheckBundleMetric) {

	// if there is nowhere to send metrics to, just return.
	if !m.check.IsReady() {
		m.Log.Printf("check not ready, skipping metric submission")
		return
	}

	// update check if there are any new metrics or, if metric tags have been added since last submit
	m.check.UpdateCheck(newMetrics)

	str, err := json.Marshal(output)
	if err != nil {
		m.Log.Printf("error preparing metrics %s", err)
		return
	}

	result, err := m.trapCall(str)
	if err != nil {
		m.Log.Printf("error sending metrics - %s\n", err)
		return
	}

	// OK response from circonus-agent does not
	// indicate how many metrics were received
	if result.Error == "agent" {
		result.Stats = uint64(len(output))
		result.Error = ""
	}

	if m.Debug || m.DumpMetrics {
		if m.DumpMetrics {
			m.Log.Printf("payload: %s", string(str))
		}
		if m.Debug {
			msg := "broker result --"
			if result.Error != "" {
				if result.Error == "agent" {
					msg += " agent does not provide response metrics"
				} else {
					msg += fmt.Sprintf(" error: %s", result.Error)
				}
			} else {
				msg += fmt.Sprintf(" stats: %d, filtered: %d", result.Stats, result.Filtered)
			}
			m.Log.Printf(msg+" duration: %s", result.Duration.String())
		}
	}
}

func (m *CirconusMetrics) trapCall(payload []byte) (*trapResult, error) {
	trap, err := m.check.GetSubmissionURL()
	if err != nil {
		return nil, errors.Wrap(err, "trap call")
	}

	dataReader := bytes.NewReader(payload)

	reqStart := time.Now()

	req, err := retryablehttp.NewRequest("PUT", trap.URL.String(), dataReader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", "cgm")
	req.Header.Set("Content-Length", strconv.Itoa(len(payload)))
	req.Close = true

	// keep last HTTP error in the event of retry failure
	var lastHTTPError error
	retryPolicy := func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return false, ctxErr
		}

		if err != nil {
			lastHTTPError = err
			return true, errors.Wrap(err, "retry policy")
		}
		// Check the response code. We retry on 500-range responses to allow
		// the server time to recover, as 500's are typically not permanent
		// errors and may relate to outages on the server side. This will catch
		// invalid response codes as well, like 0 and 999.
		if resp.StatusCode == 0 || resp.StatusCode >= 500 {
			body, readErr := ioutil.ReadAll(resp.Body)
			if readErr != nil {
				lastHTTPError = fmt.Errorf("- last HTTP error: %d %w", resp.StatusCode, readErr)
			} else {
				lastHTTPError = fmt.Errorf("- last HTTP error: %d %s", resp.StatusCode, string(body))
			}
			return true, nil
		}
		return false, nil
	}

	client := retryablehttp.NewClient()
	switch {
	case trap.URL.Scheme == "https":
		client.HTTPClient.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:       10 * time.Second,
				KeepAlive:     3 * time.Second,
				FallbackDelay: -1 * time.Millisecond,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     trap.TLS,
			DisableKeepAlives:   true,
			MaxIdleConns:        1,
			MaxIdleConnsPerHost: -1,
			DisableCompression:  false,
		}
	case trap.URL.Scheme == "http":
		client.HTTPClient.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:       10 * time.Second,
				KeepAlive:     3 * time.Second,
				FallbackDelay: -1 * time.Millisecond,
			}).DialContext,
			DisableKeepAlives:   true,
			MaxIdleConns:        1,
			MaxIdleConnsPerHost: -1,
			DisableCompression:  false,
		}
	case trap.IsSocket:
		m.Log.Printf("using socket transport\n")
		client.HTTPClient.Transport = trap.SockTransport
	default:
		return nil, fmt.Errorf("unknown scheme (%s), skipping submission", trap.URL.Scheme)
	}
	client.RetryWaitMin = 1 * time.Second
	client.RetryWaitMax = 5 * time.Second
	client.RetryMax = 3
	// retryablehttp only groks log or no log
	// but, outputs everything as [DEBUG] messages
	if m.Debug {
		client.Logger = m.Log
	} else {
		client.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
	}
	client.CheckRetry = retryPolicy

	client.ResponseLogHook = func(logger retryablehttp.Logger, r *http.Response) {
		if r.StatusCode != http.StatusOK {
			logger.Printf("non-200 response (%s): %s", r.Request.URL.String(), r.Status)
		}
	}

	attempts := -1
	client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retryNumber int) {
		if retryNumber > 0 {
			reqStart = time.Now()
			logger.Printf("retrying (%s), retry %d", req.URL.String(), retryNumber)
		}
		attempts = retryNumber
	}

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		if lastHTTPError != nil {
			return nil, fmt.Errorf("submitting: %w previous: %s attempts: %d", err, lastHTTPError, attempts)
		}
		if attempts == client.RetryMax {
			if err = m.check.RefreshTrap(); err != nil {
				return nil, fmt.Errorf("refreshing trap: %w", err)
			}
		}
		return nil, fmt.Errorf("trap call: %w", err)
	}

	dur := time.Since(reqStart)

	// no content - expected result from
	// circonus-agent when metrics accepted
	if resp.StatusCode == http.StatusNoContent {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		return &trapResult{Stats: 0, Filtered: 0, Error: "agent", Duration: dur}, nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response code: %d (%s)", resp.StatusCode, string(body))
	}

	var result trapResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing body: %w (%s)", err, body)
	}

	result.Duration = dur
	return &result, nil
}
