// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package circonusgometrics

// A Text metric is an arbitrary string
//

// SetTextWithTags sets a text metric with tags
func (m *CirconusMetrics) SetTextWithTags(metric string, tags Tags, val string) {
	m.SetTextValueWithTags(metric, tags, val)
}

// SetText sets a text metric
func (m *CirconusMetrics) SetText(metric string, val string) {
	m.SetTextValue(metric, val)
}

// SetTextValueWithTags sets a text metric with tags
func (m *CirconusMetrics) SetTextValueWithTags(metric string, tags Tags, val string) {
	m.SetTextValue(m.MetricNameWithStreamTags(metric, tags), val)
}

// SetTextValue sets a text metric
func (m *CirconusMetrics) SetTextValue(metric string, val string) {
	m.tm.Lock()
	defer m.tm.Unlock()
	m.text[metric] = val
}

// RemoveTextWithTags removes a text metric with tags
func (m *CirconusMetrics) RemoveTextWithTags(metric string, tags Tags) {
	m.RemoveText(m.MetricNameWithStreamTags(metric, tags))
}

// RemoveText removes a text metric
func (m *CirconusMetrics) RemoveText(metric string) {
	m.tm.Lock()
	defer m.tm.Unlock()
	delete(m.text, metric)
}

// SetTextFuncWithTags sets a text metric with tags to a function [called at flush interval]
func (m *CirconusMetrics) SetTextFuncWithTags(metric string, tags Tags, fn func() string) {
	m.SetTextFunc(m.MetricNameWithStreamTags(metric, tags), fn)
}

// SetTextFunc sets a text metric to a function [called at flush interval]
func (m *CirconusMetrics) SetTextFunc(metric string, fn func() string) {
	m.tfm.Lock()
	defer m.tfm.Unlock()
	m.textFuncs[metric] = fn
}

// RemoveTextFuncWithTags removes a text metric with tags function
func (m *CirconusMetrics) RemoveTextFuncWithTags(metric string, tags Tags) {
	m.RemoveTextFunc(m.MetricNameWithStreamTags(metric, tags))
}

// RemoveTextFunc a text metric function
func (m *CirconusMetrics) RemoveTextFunc(metric string) {
	m.tfm.Lock()
	defer m.tfm.Unlock()
	delete(m.textFuncs, metric)
}
