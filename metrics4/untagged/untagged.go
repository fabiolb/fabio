package untagged

import (
	"errors"
	"github.com/fabiolb/fabio/metrics4"
	"github.com/fabiolb/fabio/metrics4/prefix"
	"strings"
)

// This module provides Counter, Gauge, Timer for metric tools which don't support tags (labels)

func parseLabelsValues(labelsNames []string, labels []string) ([]string, error) {
	labelsCount := len(labelsNames)
	labelsValues := make([]string, labelsCount)

	for i := 0; i < labelsCount; i++ {
		if labelsNames[i] != labels[(i * 2)] {
			return nil, errors.New("incorrect label name")
		}

		labelsValues[i] = labels[(i * 2) + 1]
	}

	return labelsValues, nil
}

func makeNameFromLabels(labelsNames []string, labels []string) string {
	_, err := parseLabelsValues(labelsNames, labels)
	if err != nil {
		panic(err)
	}
	return strings.Join(labels, prefix.DotDelimiter)
}

type metric struct {
	p metrics4.Provider
	name string
	labelsNames []string
}

func newMetric(p metrics4.Provider, name string, labelsNames []string) *metric {
	return &metric{
		p,
		name,
		labelsNames,
	}
}

type counter struct {
	m *metric
}

func (c *counter) Add(delta float64) {}

func (c *counter) With(labels ... string) metrics4.Counter {
	return c.m.p.NewCounter(c.m.name + prefix.DotDelimiter + makeNameFromLabels(c.m.labelsNames, labels))
}

func NewCounter(p metrics4.Provider, name string, labelsNames []string) metrics4.Counter {
	return &counter{
		newMetric(p, name, labelsNames),
	}
}

type timer struct {
	m *metric
}

func (t *timer) Observe(value float64) {}

func (t *timer) With(labels ... string) metrics4.Timer {
	return t.m.p.NewTimer(t.m.name + prefix.DotDelimiter + makeNameFromLabels(t.m.labelsNames, labels))
}

func NewTimer(p metrics4.Provider, name string, labelsNames []string) metrics4.Timer {
	return &timer{
		newMetric(p, name, labelsNames),
	}
}

type gauge struct {
	m *metric
}

func (g *gauge) Add(value float64) {}

func (g *gauge) Set(value float64) {}

func (g *gauge) With(labels ... string) metrics4.Gauge {
	return g.m.p.NewGauge(g.m.name + prefix.DotDelimiter + makeNameFromLabels(g.m.labelsNames, labels))
}

func NewGauge(p metrics4.Provider, name string, labelsNames []string) metrics4.Gauge {
	return &gauge{
		newMetric(p, name, labelsNames),
	}
}

