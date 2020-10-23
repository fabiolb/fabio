// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package circonusgometrics

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/circonus-labs/circonusllhist"
	"github.com/circonus-labs/go-apiclient"
	"github.com/pkg/errors"
)

func (m *CirconusMetrics) packageMetrics() (map[string]*apiclient.CheckBundleMetric, Metrics) {

	m.packagingmu.Lock()
	defer m.packagingmu.Unlock()

	if m.Debug {
		m.Log.Printf("packaging metrics\n")
	}

	counters, gauges, histograms, text := m.snapshot()
	newMetrics := make(map[string]*apiclient.CheckBundleMetric)
	output := make(Metrics, len(counters)+len(gauges)+len(histograms)+len(text))
	for name, value := range counters {
		send := m.check.IsMetricActive(name)
		if !send && m.check.ActivateMetric(name) {
			send = true
			newMetrics[name] = &apiclient.CheckBundleMetric{
				Name:   name,
				Type:   "numeric",
				Status: "active",
			}
		}
		if send {
			output[name] = Metric{Type: "L", Value: value}
		}
	}

	for name, value := range gauges {
		send := m.check.IsMetricActive(name)
		if !send && m.check.ActivateMetric(name) {
			send = true
			newMetrics[name] = &apiclient.CheckBundleMetric{
				Name:   name,
				Type:   "numeric",
				Status: "active",
			}
		}
		if send {
			output[name] = Metric{Type: m.getGaugeType(value), Value: value}
		}
	}

	for name, value := range histograms {
		send := m.check.IsMetricActive(name)
		if !send && m.check.ActivateMetric(name) {
			send = true
			newMetrics[name] = &apiclient.CheckBundleMetric{
				Name:   name,
				Type:   "histogram",
				Status: "active",
			}
		}
		if send {
			output[name] = Metric{Type: "h", Value: value.DecStrings()}
		}
	}

	for name, value := range text {
		send := m.check.IsMetricActive(name)
		if !send && m.check.ActivateMetric(name) {
			send = true
			newMetrics[name] = &apiclient.CheckBundleMetric{
				Name:   name,
				Type:   "text",
				Status: "active",
			}
		}
		if send {
			output[name] = Metric{Type: "s", Value: value}
		}
	}

	m.lastMetrics.metricsmu.Lock()
	defer m.lastMetrics.metricsmu.Unlock()
	m.lastMetrics.metrics = &output
	m.lastMetrics.ts = time.Now()

	return newMetrics, output
}

// PromOutput returns lines of metrics in prom format
func (m *CirconusMetrics) PromOutput() (*bytes.Buffer, error) {
	m.lastMetrics.metricsmu.Lock()
	defer m.lastMetrics.metricsmu.Unlock()

	if m.lastMetrics.metrics == nil {
		return nil, errors.New("no metrics available")
	}

	var b bytes.Buffer
	w := bufio.NewWriter(&b)

	ts := m.lastMetrics.ts.UnixNano() / int64(time.Millisecond)

	for name, metric := range *m.lastMetrics.metrics {
		switch metric.Type {
		case "n":
			if strings.HasPrefix(fmt.Sprintf("%v", metric.Value), "[H[") {
				continue // circonus histogram != prom "histogram" (aka percentile)
			}
		case "h":
			continue // circonus histogram != prom "histogram" (aka percentile)
		case "s":
			continue // text metrics unsupported
		}
		fmt.Fprintf(w, "%s %v %d\n", name, metric.Value, ts)
	}

	err := w.Flush()
	if err != nil {
		return nil, errors.Wrap(err, "flushing metric buffer")
	}

	return &b, err
}

// FlushMetricsNoReset flushes current metrics to a structure and returns it (does NOT send to Circonus).
func (m *CirconusMetrics) FlushMetricsNoReset() *Metrics {
	m.flushmu.Lock()
	if m.flushing {
		m.flushmu.Unlock()
		return &Metrics{}
	}

	m.flushing = true
	m.flushmu.Unlock()

	// save values configured at startup
	resetC := m.resetCounters
	resetG := m.resetGauges
	resetH := m.resetHistograms
	resetT := m.resetText
	// override Reset* to false for this call
	m.resetCounters = false
	m.resetGauges = false
	m.resetHistograms = false
	m.resetText = false

	_, output := m.packageMetrics()

	// restore previous values
	m.resetCounters = resetC
	m.resetGauges = resetG
	m.resetHistograms = resetH
	m.resetText = resetT

	m.flushmu.Lock()
	m.flushing = false
	m.flushmu.Unlock()

	return &output
}

// FlushMetrics flushes current metrics to a structure and returns it (does NOT send to Circonus)
func (m *CirconusMetrics) FlushMetrics() *Metrics {
	m.flushmu.Lock()
	if m.flushing {
		m.flushmu.Unlock()
		return &Metrics{}
	}

	m.flushing = true
	m.flushmu.Unlock()

	_, output := m.packageMetrics()

	m.flushmu.Lock()
	m.flushing = false
	m.flushmu.Unlock()

	return &output
}

// Flush metrics kicks off the process of sending metrics to Circonus
func (m *CirconusMetrics) Flush() {
	m.flushmu.Lock()
	if m.flushing {
		m.flushmu.Unlock()
		return
	}

	m.flushing = true
	m.flushmu.Unlock()

	newMetrics, output := m.packageMetrics()

	if len(output) > 0 {
		m.submit(output, newMetrics)
	} else if m.Debug {
		m.Log.Printf("no metrics to send, skipping\n")
	}

	m.flushmu.Lock()
	m.flushing = false
	m.flushmu.Unlock()
}

// Reset removes all existing counters and gauges.
func (m *CirconusMetrics) Reset() {
	m.cm.Lock()
	defer m.cm.Unlock()

	m.cfm.Lock()
	defer m.cfm.Unlock()

	m.gm.Lock()
	defer m.gm.Unlock()

	m.gfm.Lock()
	defer m.gfm.Unlock()

	m.hm.Lock()
	defer m.hm.Unlock()

	m.tm.Lock()
	defer m.tm.Unlock()

	m.tfm.Lock()
	defer m.tfm.Unlock()

	m.counters = make(map[string]uint64)
	m.counterFuncs = make(map[string]func() uint64)
	m.gauges = make(map[string]interface{})
	m.gaugeFuncs = make(map[string]func() int64)
	m.histograms = make(map[string]*Histogram)
	m.text = make(map[string]string)
	m.textFuncs = make(map[string]func() string)
}

// snapshot returns a copy of the values of all registered counters and gauges.
func (m *CirconusMetrics) snapshot() (c map[string]uint64, g map[string]interface{}, h map[string]*circonusllhist.Histogram, t map[string]string) {
	c = m.snapCounters()
	g = m.snapGauges()
	h = m.snapHistograms()
	t = m.snapText()

	return
}

func (m *CirconusMetrics) snapCounters() map[string]uint64 {
	m.cm.Lock()
	defer m.cm.Unlock()
	m.cfm.Lock()
	defer m.cfm.Unlock()

	c := make(map[string]uint64, len(m.counters)+len(m.counterFuncs))

	for n, v := range m.counters {
		c[n] = v
	}
	if m.resetCounters && len(c) > 0 {
		m.counters = make(map[string]uint64)
	}

	for n, f := range m.counterFuncs {
		c[n] = f()
	}

	return c
}

func (m *CirconusMetrics) snapGauges() map[string]interface{} {
	m.gm.Lock()
	defer m.gm.Unlock()
	m.gfm.Lock()
	defer m.gfm.Unlock()

	g := make(map[string]interface{}, len(m.gauges)+len(m.gaugeFuncs))

	for n, v := range m.gauges {
		g[n] = v
	}
	if m.resetGauges && len(g) > 0 {
		m.gauges = make(map[string]interface{})
	}

	for n, f := range m.gaugeFuncs {
		g[n] = f()
	}

	return g
}

func (m *CirconusMetrics) snapHistograms() map[string]*circonusllhist.Histogram {
	m.hm.Lock()
	defer m.hm.Unlock()

	h := make(map[string]*circonusllhist.Histogram, len(m.histograms))

	for n, hist := range m.histograms {
		hist.rw.Lock()
		if m.resetHistograms {
			h[n] = hist.hist.CopyAndReset()
		} else {
			h[n] = hist.hist.Copy()
		}

		hist.rw.Unlock()
	}

	if m.resetHistograms && len(h) > 0 {
		m.histograms = make(map[string]*Histogram)
	}

	return h
}

func (m *CirconusMetrics) snapText() map[string]string {
	m.tm.Lock()
	defer m.tm.Unlock()
	m.tfm.Lock()
	defer m.tfm.Unlock()

	t := make(map[string]string, len(m.text)+len(m.textFuncs))

	for n, v := range m.text {
		t[n] = v
	}
	if m.resetText && len(t) > 0 {
		m.text = make(map[string]string)
	}

	for n, f := range m.textFuncs {
		t[n] = f()
	}

	return t
}
