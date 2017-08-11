package noop

import "time"

type Registry struct{}

func (p Registry) Names(group string) []string              { return nil }
func (p Registry) Unregister(group, name string)            {}
func (p Registry) UnregisterAll(group string)               {}
func (p Registry) Gauge(group, name string, n float64)      {}
func (p Registry) Inc(group, name string, n int64)          {}
func (p Registry) Time(group, name string, d time.Duration) {}
