package route

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// picker selects a target from a list of targets.
type picker func(r *Route) *Target

// Picker contains the available picker functions.
// Update config/load.go#load after updating.
var Picker = map[string]picker{
	"rnd": rndPicker,
	"rr":  rrPicker,
}

// rndPicker picks a random target from the list of targets.
func rndPicker(r *Route) *Target {
	return r.wTargets[randIntn(len(r.wTargets))]
}

// rrPicker picks the next target from a list of targets using round-robin.
func rrPicker(r *Route) *Target {
	u := r.wTargets[r.total%uint64(len(r.wTargets))]
	atomic.AddUint64(&r.total, 1)
	return u
}

// as it turns out, math/rand's Intn is now way faster (4x) than the previous implementation using
// time.UnixNano().  As a bonus, this actually works properly on 32 bit platforms.
var rndOnce sync.Once
var randIntn = func(n int) int {
	rndOnce.Do(func() {
		rand.Seed(time.Now().UnixNano())
	})
	if n == 0 {
		return 0
	}
	return rand.Intn(n)
}
