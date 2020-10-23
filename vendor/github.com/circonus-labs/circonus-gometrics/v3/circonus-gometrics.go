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
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/circonus-labs/circonus-gometrics/v3/checkmgr"
	"github.com/pkg/errors"
)

const (
	defaultFlushInterval = "10s" // 10 * time.Second
)

// Logger facilitates use of any logger supporting the required methods
// rather than just standard log package log.Logger
type Logger interface {
	Printf(string, ...interface{})
}

// Metric defines an individual metric
type Metric struct {
	Type  string      `json:"_type"`
	Value interface{} `json:"_value"`
}

// Metrics holds host metrics
type Metrics map[string]Metric

// Config options for circonus-gometrics
type Config struct {
	Log             Logger
	Debug           bool
	ResetCounters   string // reset/delete counters on flush (default true)
	ResetGauges     string // reset/delete gauges on flush (default true)
	ResetHistograms string // reset/delete histograms on flush (default true)
	ResetText       string // reset/delete text on flush (default true)

	// API, Check and Broker configuration options
	CheckManager checkmgr.Config

	// how frequenly to submit metrics to Circonus, default 10 seconds.
	// Set to 0 to disable automatic flushes and call Flush manually.
	Interval string
}

type prevMetrics struct {
	metrics   *Metrics
	metricsmu sync.Mutex
	ts        time.Time
}

// CirconusMetrics state
type CirconusMetrics struct {
	Log   Logger
	Debug bool

	resetCounters   bool
	resetGauges     bool
	resetHistograms bool
	resetText       bool
	flushInterval   time.Duration
	flushing        bool
	flushmu         sync.Mutex
	packagingmu     sync.Mutex
	check           *checkmgr.CheckManager
	lastMetrics     *prevMetrics

	counters map[string]uint64
	cm       sync.Mutex

	counterFuncs map[string]func() uint64
	cfm          sync.Mutex

	gauges map[string]interface{}
	gm     sync.Mutex

	gaugeFuncs map[string]func() int64
	gfm        sync.Mutex

	histograms map[string]*Histogram
	hm         sync.Mutex

	text map[string]string
	tm   sync.Mutex

	textFuncs map[string]func() string
	tfm       sync.Mutex
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
		lastMetrics:  &prevMetrics{},
	}

	// Logging
	{
		cm.Debug = cfg.Debug
		cm.Log = cfg.Log

		if cm.Debug && cm.Log == nil {
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

	// start background initialization
	cm.check.Initialize()

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
