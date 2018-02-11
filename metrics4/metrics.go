package metrics4

import (
	"time"
)

type Provider interface {
	NewCounter(name string, labels ...string) Counter
	NewTimer(name string, labels ...string) Timer
	Unregister(v interface{})
}

type MultiProvider struct {
	p []Provider
}

func (mp *MultiProvider) Register(p Provider) {
	mp.p = append(mp.p, p)
}

func (mp *MultiProvider) NewCounter(name string, labels ...string) Counter {
	m := &MultiCounter{}
	for _, p := range mp.p {
		m.Register(p.NewCounter(name, labels...))
	}
	return m
}

func (mp *MultiProvider) NewTimer(name string, labels ...string) Timer {
	m := &MultiTimer{}
	for _, p := range mp.p {
		m.Register(p.NewTimer(name, labels...))
	}
	return m
}

func (mp *MultiProvider) Unregister(v interface{}) {
	for _, p := range mp.p {
		p.Unregister(v)
	}
}

type Counter interface {
	Count(int)
}

type MultiCounter struct {
	c []Counter
}

func (mc *MultiCounter) Register(c Counter) {
	mc.c = append(mc.c, c)
}

func (mc *MultiCounter) Count(n int) {
	for _, c := range mc.c {
		c.Count(n)
	}
}

type Timer interface {
	Update(time.Duration)
}

type MultiTimer struct {
	t []Timer
}

func (mt *MultiTimer) Register(t Timer) {
	mt.t = append(mt.t, t)
}

func (mt *MultiTimer) Update(d time.Duration) {
	for _, t := range mt.t {
		t.Update(d)
	}
}
