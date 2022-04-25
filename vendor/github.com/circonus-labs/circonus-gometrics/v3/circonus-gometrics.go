// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package circonusgometrics provides instrumentation for your applications in the form
// of counters, gauges and histograms and allows you to publish them to
// Circonus
//
// Counters
//
// A counter is a monotonically-increasing, unsigned, 64-bit integer used to
// represent the number of times an event has occurred. By tracking the deltas
// between measurements of a counter over intervals of time, an aggregation
// layer can derive rates, acceleration, etc.
//
// Gauges
//
// A gauge returns instantaneous measurements of something using signed, 64-bit
// integers. This value does not need to be monotonic.
//
// Histograms
//
// A histogram tracks the distribution of a stream of values (e.g. the number of
// seconds it takes to handle requests).  Circonus can calculate complex
// analytics on these.
//
// Reporting
//
// A period push to a Circonus httptrap is confgurable.
package circonusgometrics

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/circonus-labs/circonus-gometrics/v3/checkmgr"
	"github.com/circonus-labs/go-apiclient"
	"github.com/pkg/errors"
)

const (
	defaultFlushInterval = "10s" // 10 * time.Second

	// MetricTypeInt32 reconnoiter
	MetricTypeInt32 = "i"

	// MetricTypeUint32 reconnoiter
	MetricTypeUint32 = "I"

	// MetricTypeInt64 reconnoiter
	MetricTypeInt64 = "l"

	// MetricTypeUint64 reconnoiter
	MetricTypeUint64 = "L"

	// MetricTypeFloat64 reconnoiter
	MetricTypeFloat64 = "n"

	// MetricTypeString reconnoiter
	MetricTypeString = "s"

	// MetricTypeHistogram reconnoiter
	MetricTypeHistogram = "h"

	// MetricTypeCumulativeHistogram reconnoiter
	MetricTypeCumulativeHistogram = "H"
)

var (
	metricTypeRx = regexp.MustCompile(`^[` + strings.Join([]string{
		MetricTypeInt32,
		MetricTypeUint32,
		MetricTypeInt64,
		MetricTypeUint64,
		MetricTypeFloat64,
		MetricTypeString,
		MetricTypeHistogram,
		MetricTypeCumulativeHistogram,
	}, "") + `]$`)
)

// Logger facilitates use of any logger supporting the required methods
// rather than just standard log package log.Logger
type Logger interface {
	Printf(string, ...interface{})
}

// Metric defines an individual metric
type Metric struct {
	Value     interface{} `json:"_value"`
	Type      string      `json:"_type"`
	Timestamp uint64      `json:"_ts,omitempty"`
}

// Metrics holds host metrics
type Metrics map[string]Metric

// Config options for circonus-gometrics
type Config struct {
	Log             Logger
	ResetCounters   string // reset/delete counters on flush (default true)
	ResetGauges     string // reset/delete gauges on flush (default true)
	ResetHistograms string // reset/delete histograms on flush (default true)
	ResetText       string // reset/delete text on flush (default true)
	// how frequenly to submit metrics to Circonus, default 10 seconds.
	// Set to 0 to disable automatic flushes and call Flush manually.
	Interval string

	// API, Check and Broker configuration options
	CheckManager checkmgr.Config

	Debug       bool
	DumpMetrics bool
}

type prevMetrics struct {
	ts        time.Time
	metricsmu sync.Mutex
	metrics   *Metrics
}

// CirconusMetrics state
type CirconusMetrics struct {
	Log             Logger
	lastMetrics     *prevMetrics
	check           *checkmgr.CheckManager
	gauges          map[string]interface{}
	histograms      map[string]*Histogram
	custom          map[string]Metric
	text            map[string]string
	textFuncs       map[string]func() string
	counterFuncs    map[string]func() uint64
	gaugeFuncs      map[string]func() int64
	counters        map[string]uint64
	submitTimestamp *time.Time
	flushInterval   time.Duration
	flushmu         sync.Mutex
	packagingmu     sync.Mutex
	cm              sync.Mutex
	cfm             sync.Mutex
	gm              sync.Mutex
	gfm             sync.Mutex
	hm              sync.Mutex
	tm              sync.Mutex
	tfm             sync.Mutex
	custm           sync.Mutex
	flushing        bool
	Debug           bool
	DumpMetrics     bool
	resetCounters   bool
	resetGauges     bool
	resetHistograms bool
	resetText       bool
}

// NewCirconusMetrics returns a CirconusMetrics instance
func NewCirconusMetrics(cfg *Config) (*CirconusMetrics, error) {
	return New(cfg)
}

// New returns a CirconusMetrics instance
func New(cfg *Config) (*CirconusMetrics, error) {

	if cfg == nil {
		return nil, errors.New("invalid configuration (nil)")
	}

	cm := &CirconusMetrics{
		counters:     make(map[string]uint64),
		counterFuncs: make(map[string]func() uint64),
		gauges:       make(map[string]interface{}),
		gaugeFuncs:   make(map[string]func() int64),
		histograms:   make(map[string]*Histogram),
		text:         make(map[string]string),
		textFuncs:    make(map[string]func() string),
		custom:       make(map[string]Metric),
		lastMetrics:  &prevMetrics{},
	}

	// Logging
	{
		cm.Debug = cfg.Debug
		cm.DumpMetrics = cfg.DumpMetrics
		cm.Log = cfg.Log

		if (cm.Debug || cm.DumpMetrics) && cm.Log == nil {
			cm.Log = log.New(os.Stderr, "", log.LstdFlags)
		}
		if cm.Log == nil {
			cm.Log = log.New(ioutil.Discard, "", log.LstdFlags)
		}
	}

	// Flush Interval
	{
		fi := defaultFlushInterval
		if cfg.Interval != "" {
			fi = cfg.Interval
		}

		dur, err := time.ParseDuration(fi)
		if err != nil {
			return nil, errors.Wrap(err, "parsing flush interval")
		}
		cm.flushInterval = dur
	}

	// metric resets

	cm.resetCounters = true
	if cfg.ResetCounters != "" {
		setting, err := strconv.ParseBool(cfg.ResetCounters)
		if err != nil {
			return nil, errors.Wrap(err, "parsing reset counters")
		}
		cm.resetCounters = setting
	}

	cm.resetGauges = true
	if cfg.ResetGauges != "" {
		setting, err := strconv.ParseBool(cfg.ResetGauges)
		if err != nil {
			return nil, errors.Wrap(err, "parsing reset gauges")
		}
		cm.resetGauges = setting
	}

	cm.resetHistograms = true
	if cfg.ResetHistograms != "" {
		setting, err := strconv.ParseBool(cfg.ResetHistograms)
		if err != nil {
			return nil, errors.Wrap(err, "parsing reset histograms")
		}
		cm.resetHistograms = setting
	}

	cm.resetText = true
	if cfg.ResetText != "" {
		setting, err := strconv.ParseBool(cfg.ResetText)
		if err != nil {
			return nil, errors.Wrap(err, "parsing reset text")
		}
		cm.resetText = setting
	}

	// check manager
	{
		cfg.CheckManager.Debug = cm.Debug
		cfg.CheckManager.Log = cm.Log

		check, err := checkmgr.New(&cfg.CheckManager)
		if err != nil {
			return nil, errors.Wrap(err, "creating new check manager")
		}
		cm.check = check
	}

	// start initialization (serialized or background)
	if err := cm.check.Initialize(); err != nil {
		return nil, err
	}

	// if automatic flush is enabled, start it.
	// NOTE: submit will jettison metrics until initialization has completed.
	if cm.flushInterval > time.Duration(0) {
		go func() {
			for range time.NewTicker(cm.flushInterval).C {
				cm.Flush()
			}
		}()
	}

	return cm, nil
}

// Start deprecated NOP, automatic flush is started in New if flush interval > 0.
func (m *CirconusMetrics) Start() {
	// nop
}

// Ready returns true or false indicating if the check is ready to accept metrics
func (m *CirconusMetrics) Ready() bool {
	return m.check.IsReady()
}

// Custom adds a user defined metric
func (m *CirconusMetrics) Custom(metricName string, metric Metric) error {
	if !metricTypeRx.MatchString(metric.Type) {
		return fmt.Errorf("unrecognized circonus metric type (%s)", metric.Type)
	}

	m.custm.Lock()
	m.custom[metricName] = metric
	m.custm.Unlock()

	return nil
}

// GetBrokerTLSConfig returns the tls.Config for the broker
func (m *CirconusMetrics) GetBrokerTLSConfig() *tls.Config {
	return m.check.BrokerTLSConfig()
}

func (m *CirconusMetrics) GetCheckBundle() *apiclient.CheckBundle {
	return m.check.GetCheckBundle()
}

func (m *CirconusMetrics) SetSubmitTimestamp(ts time.Time) {
	m.packagingmu.Lock()
	defer m.packagingmu.Unlock()
	m.submitTimestamp = &ts
}
