// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package circonusgometrics

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/circonus-labs/go-apiclient"
	"github.com/openhistogram/circonusllhist"
	"github.com/pkg/errors"
)

func (m *CirconusMetrics) packageMetrics() (map[string]*apiclient.CheckBundleMetric, Metrics) {

	m.packagingmu.Lock()
	defer m.packagingmu.Unlock()

	// if m.Debug {
	// 	m.Log.Printf("packaging metrics\n")
	// }

	var ts uint64
	// always submitting a timestamp forces the broker to treat the check as though
	// it is "async" which doesn't work well for "group" checks with multiple submitters
	// e.g. circonus-agent with a group statsd check
	// if m.submitTimestamp == nil {
	// 	ts = makeTimestamp(time.Now())
	// } else {
	if m.submitTimestamp != nil {
		ts = makeTimestamp(*m.submitTimestamp)
		m.Log.Printf("setting custom timestamp %v -> %v (UTC ms)", *m.submitTimestamp, ts)
	}

	newMetrics := make(map[string]*apiclient.CheckBundleMetric)
	counters, gauges, histograms, text := m.snapshot()
	m.custm.Lock()
	output := make(Metrics, len(counters)+len(gauges)+len(histograms)+len(text)+len(m.custom))
	if len(m.custom) > 0 {
		// add and reset any custom metrics
		for mn, mv := range m.custom {
			output[mn] = mv
		}
		m.custom = make(map[string]Metric)
	}
	m.custm.Unlock()
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
			metric := Metric{Type: "L", Value: value}
			if ts > 0 {
				metric.Timestamp = ts
			}
			output[name] = metric
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
			metric := Metric{Type: m.getGaugeType(value), Value: value}
			if ts > 0 {
				metric.Timestamp = ts
			}
			output[name] = metric
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
			buf := bytes.NewBuffer([]byte{})
			if err := value.SerializeB64(buf); err != nil {
				m.Log.Printf("[ERR] serializing histogram %s: %s", name, err)
			} else {
				// histograms b64 serialized support timestamps
				metric := Metric{Type: "h", Value: buf.String()}
				if ts > 0 {
					metric.Timestamp = ts
				}
				output[name] = metric
			}
			// output[name] = Metric{Type: "h", Value: value.DecStrings()} // histograms do NOT get timestamps
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
			metric := Metric{Type: "s", Value: value}
			if ts > 0 {
				metric.Timestamp = ts
			}
			output[name] = metric
		}
	}

	m.lastMetrics.metricsmu.Lock()
	defer m.lastMetrics.metricsmu.Unlock()
	m.lastMetrics.metrics = &output
	m.lastMetrics.ts = time.Now()
	// reset the submission timestamp
	m.submitTimestamp = nil

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
	} /* else if m.Debug {
		m.Log.Printf("no metrics to send, skipping\n")
	}*/

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
func (m *CirconusMetrics) snapshot() (
	map[string]uint64, // counters
	map[string]interface{}, // gauges
	map[string]*circonusllhist.Histogram, // histograms
	map[string]string) { // text

	var h map[string]*circonusllhist.Histogram
	var c map[string]uint64
	var g map[string]interface{}
	var t map[string]string

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		h = m.snapHistograms()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		c = m.snapCounters()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		g = m.snapGauges()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		t = m.snapText()
		wg.Done()
	}()

	wg.Wait()

	return c, g, h, t
}

func (m *CirconusMetrics) snapCounters() map[string]uint64 {
	m.cm.Lock()
	m.cfm.Lock()

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

	m.cm.Unlock()
	m.cfm.Unlock()

	return c
}

func (m *CirconusMetrics) snapGauges() map[string]interface{} {
	m.gm.Lock()
	m.gfm.Lock()

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

	m.gm.Unlock()
	m.gfm.Unlock()

	return g
}

func (m *CirconusMetrics) snapHistograms() map[string]*circonusllhist.Histogram {
	m.hm.Lock()

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

	m.hm.Unlock()

	return h
}

func (m *CirconusMetrics) snapText() map[string]string {
	m.tm.Lock()
	m.tfm.Lock()

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

	m.tm.Unlock()
	m.tfm.Unlock()

	return t
}

// makeTimestamp returns timestamp in ms units for _ts metric value
func makeTimestamp(ts time.Time) uint64 {
	return uint64(ts.UTC().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond)))
}
