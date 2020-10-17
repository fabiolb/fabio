// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package circonusgometrics

import "github.com/pkg/errors"

// A Counter is a monotonically increasing unsigned integer.
//
// Use a counter to derive rates (e.g., record total number of requests, derive
// requests per second).

// IncrementWithTags counter by 1, with tags
func (m *CirconusMetrics) IncrementWithTags(metric string, tags Tags) {
	m.AddWithTags(metric, tags, 1)
}

// Increment counter by 1
func (m *CirconusMetrics) Increment(metric string) {
	m.Add(metric, 1)
}

// IncrementByValueWithTags updates counter metric with tags by supplied value
func (m *CirconusMetrics) IncrementByValueWithTags(metric string, tags Tags, val uint64) {
	m.AddWithTags(metric, tags, val)
}

// IncrementByValue updates counter by supplied value
func (m *CirconusMetrics) IncrementByValue(metric string, val uint64) {
	m.Add(metric, val)
}

// SetWithTags sets a counter metric with tags to specific value
func (m *CirconusMetrics) SetWithTags(metric string, tags Tags, val uint64) {
	m.Set(m.MetricNameWithStreamTags(metric, tags), val)
}

// Set a counter to specific value
func (m *CirconusMetrics) Set(metric string, val uint64) {
	m.cm.Lock()
	defer m.cm.Unlock()
	m.counters[metric] = val
}

// AddWithTags updates counter metric with tags by supplied value
func (m *CirconusMetrics) AddWithTags(metric string, tags Tags, val uint64) {
	m.Add(m.MetricNameWithStreamTags(metric, tags), val)
}

// Add updates counter by supplied value
func (m *CirconusMetrics) Add(metric string, val uint64) {
	m.cm.Lock()
	defer m.cm.Unlock()
	m.counters[metric] += val
}

// RemoveCounterWithTags removes the named counter metric with tags
func (m *CirconusMetrics) RemoveCounterWithTags(metric string, tags Tags) {
	m.RemoveCounter(m.MetricNameWithStreamTags(metric, tags))
}

// RemoveCounter removes the named counter
func (m *CirconusMetrics) RemoveCounter(metric string) {
	m.cm.Lock()
	defer m.cm.Unlock()
	delete(m.counters, metric)
}

// GetCounterTest returns the current value for a counter. (note: it is a function specifically for "testing", disable automatic submission during testing.)
func (m *CirconusMetrics) GetCounterTest(metric string) (uint64, error) {
	m.cm.Lock()
	defer m.cm.Unlock()

	if val, ok := m.counters[metric]; ok {
		return val, nil
	}

	return 0, errors.Errorf("counter metric '%s' not found", metric)

}

// SetCounterFuncWithTags set counter metric with tags to a function [called at flush interval]
func (m *CirconusMetrics) SetCounterFuncWithTags(metric string, tags Tags, fn func() uint64) {
	m.SetCounterFunc(m.MetricNameWithStreamTags(metric, tags), fn)
}

// SetCounterFunc set counter to a function [called at flush interval]
func (m *CirconusMetrics) SetCounterFunc(metric string, fn func() uint64) {
	m.cfm.Lock()
	defer m.cfm.Unlock()
	m.counterFuncs[metric] = fn
}

// RemoveCounterFuncWithTags removes the named counter metric function with tags
func (m *CirconusMetrics) RemoveCounterFuncWithTags(metric string, tags Tags) {
	m.RemoveCounterFunc(m.MetricNameWithStreamTags(metric, tags))
}

// RemoveCounterFunc removes the named counter function
func (m *CirconusMetrics) RemoveCounterFunc(metric string) {
	m.cfm.Lock()
	defer m.cfm.Unlock()
	delete(m.counterFuncs, metric)
}
