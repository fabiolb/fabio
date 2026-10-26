package route

import (
	"math/rand"
	"sync/atomic"
)

// picker selects a target from a list of targets.
type picker func(r *Route) *Target

// Picker contains the available picker functions.
// Update config/load.go#load after updating.
var Picker = map[string]picker{
	"rnd":   rndPicker,
	"rr":    rrPicker,
	"lconn": lconnPicker,
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

// lconnPicker first honors the weighted allocation, then chooses the least
// active target within that allocation. Ties rotate among distinct targets.
func lconnPicker(r *Route) *Target {
	if len(r.wTargets) == 0 {
		return nil
	}
	group := r.wTargets[randIntn(len(r.wTargets))].weightGroup
	var best *Target
	var lowest int64
	ties := 0
	for _, t := range r.Targets {
		if t.weightGroup != group || t.Weight <= 0 {
			continue
		}
		active := t.Active()
		if best == nil || active < lowest {
			best = t
			lowest = active
			ties = 1
		} else if active == lowest {
			ties++
		}
	}
	if ties <= 1 {
		return best
	}
	// Each allocation rotates its own ties, independent of other groups.
	index := (group.Add(1) - 1) % uint64(ties)
	for _, t := range r.Targets {
		if t.weightGroup == group && t.Weight > 0 && t.Active() == lowest {
			if index == 0 {
				return t
			}
			index--
		}
	}
	return best
}

// as it turns out, math/rand's Intn is now way faster (4x) than the previous implementation using
// time.UnixNano().  As a bonus, this actually works properly on 32 bit platforms.
var randIntn = func(n int) int {
	if n == 0 {
		return 0
	}
	return rand.Intn(n)
}
