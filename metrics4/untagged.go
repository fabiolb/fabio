package metrics4

import (
	"errors"
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
	return strings.Join(labels, "_")
}

type untaggedMetric struct {
	p Provider
	name string
	labelsNames []string
}

func newUntaggedMetric(p Provider, name string, labelsNames []string) *untaggedMetric {
	return &untaggedMetric{
		p,
		name,
		labelsNames,
	}
}

type untaggedCounter struct {
	m *untaggedMetric
}

func (c *untaggedCounter) Add(delta float64) {}

func (c *untaggedCounter) With(labels ... string) Counter {
	return c.m.p.NewCounter(c.m.name + "_" + makeNameFromLabels(c.m.labelsNames, labels))
}

func NewUntaggedCounter(p Provider, name string, labelsNames []string) Counter {
	return &untaggedCounter{
		newUntaggedMetric(p, name, labelsNames),
	}
}

type untaggedTimer struct {
	m *untaggedMetric
}

func (t *untaggedTimer) Observe(value float64) {}

func (t *untaggedTimer) With(labels ... string) Timer {
	return t.m.p.NewTimer(t.m.name + "_" + makeNameFromLabels(t.m.labelsNames, labels))
}

func NewUntaggedTimer(p Provider, name string, labelsNames []string) Timer {
	return &untaggedTimer{
		newUntaggedMetric(p, name, labelsNames),
	}
}

type untaggedGauge struct {
	m *untaggedMetric
}

func (g *untaggedGauge) Add(value float64) {}

func (g *untaggedGauge) Set(value float64) {}

func (g *untaggedGauge) With(labels ... string) Gauge {
	return g.m.p.NewGauge(g.m.name + "_" + makeNameFromLabels(g.m.labelsNames, labels))
}

func NewUntaggedGauge(p Provider, name string, labelsNames []string) Gauge {
	return &untaggedGauge{
		newUntaggedMetric(p, name, labelsNames),
	}
}

