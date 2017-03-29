package route

import (
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

// stubbed out for testing
// we implement the randIntN function using the nanosecond time counter
// since it is 15x faster than using the pseudo random number generator
// (12 ns vs 190 ns) Most HW does not seem to provide clocks with ns
// resolution but seem to be good enough for µs resolution. Since
// requests are usually handled within several ms we should have enough
// variation. Within 1 ms we have 1000 µs to distribute among a smaller
// set of entities (<< 100)
var randIntn = func(n int) int {
	if n == 0 {
		return 0
	}
	return int(time.Now().UnixNano()/int64(time.Microsecond)) % n
}
