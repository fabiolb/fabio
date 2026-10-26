package route

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
)

func leastConnTable(t *testing.T, lines string) Table {
	t.Helper()
	tbl, err := NewTable(bytes.NewBufferString(lines))
	if err != nil {
		t.Fatal(err)
	}
	return tbl
}

func TestLeastConnPicker(t *testing.T) {
	r := &Route{Host: "", Path: "/"}
	r.addTarget("svc", fooDotCom, -1, nil, nil)
	r.addTarget("svc", barDotCom, -1, nil, nil)

	for i := range 6 {
		if got, want := lconnPicker(r), r.Targets[i%2]; got != want {
			t.Fatalf("tie %d: got %v, want %v", i, got.URL, want.URL)
		}
	}

	r.Targets[0].Begin()
	if got := lconnPicker(r); got != r.Targets[1] {
		t.Fatalf("busy first target: got %v, want %v", got.URL, r.Targets[1].URL)
	}
	r.Targets[1].Begin()
	r.Targets[1].Begin()
	if got := lconnPicker(r); got != r.Targets[0] {
		t.Fatalf("busy second target: got %v, want %v", got.URL, r.Targets[0].URL)
	}
	r.Targets[0].End()
	r.Targets[1].End()
	r.Targets[1].End()
	for _, target := range r.Targets {
		if got := target.Active(); got != 0 {
			t.Errorf("%s: %d active operations remain", target.URL, got)
		}
	}
}

func TestLeastConnWeightedGroups(t *testing.T) {
	tbl := leastConnTable(t, `route add stable / http://main-a/
route add stable / http://main-b/
route add canary / http://canary-west/ tags "canary,west"
route add canary / http://canary-east/ tags "canary,east"
route weight / weight 0.05 tags "canary"`)
	r := tbl.route("", "/")
	if r.Targets[2].weightGroup != r.Targets[3].weightGroup || r.Targets[0].weightGroup == r.Targets[2].weightGroup {
		t.Fatal("weighted canary targets must share a separate allocation")
	}

	canarySlots := 0
	canaryIndex, stableIndex := -1, -1
	for i, target := range r.wTargets {
		if target.Service == "canary" {
			canarySlots++
			canaryIndex = i
		} else {
			stableIndex = i
		}
	}
	if got, want := canarySlots, len(r.wTargets)/20; got != want {
		t.Fatalf("canary allocation: %d/%d slots, want %d", got, len(r.wTargets), want)
	}

	previous := randIntn
	defer func() { randIntn = previous }()
	r.Targets[2].Begin()
	r.Targets[0].Begin()
	randIntn = func(int) int { return canaryIndex }
	if got := lconnPicker(r); got != r.Targets[3] {
		t.Fatalf("canary group: got %s, want %s", got.URL, r.Targets[3].URL)
	}
	randIntn = func(int) int { return stableIndex }
	if got := lconnPicker(r); got != r.Targets[1] {
		t.Fatalf("stable group: got %s, want %s", got.URL, r.Targets[1].URL)
	}
	r.Targets[2].End()
	r.Targets[0].End()

	for i := range 6 {
		randIntn = func(int) int { return canaryIndex }
		if got := lconnPicker(r); got != r.Targets[2+i%2] {
			t.Fatalf("canary tie %d: got %s", i, got.URL)
		}
		randIntn = func(int) int { return stableIndex }
		if got := lconnPicker(r); got != r.Targets[i%2] {
			t.Fatalf("stable tie %d: got %s", i, got.URL)
		}
	}
}

func TestLeastConnFixedWeightsAndZero(t *testing.T) {
	tbl := leastConnTable(t, `route add stable / http://main/ weight 0.5
route add canary / http://canary/ weight 0.5
route add disabled / http://disabled/ weight 0`)
	r := tbl.route("", "/")
	previous := randIntn
	defer func() { randIntn = previous }()
	r.Targets[0].Begin()
	r.Targets[0].Begin()
	r.Targets[1].Begin()
	for i := range r.wTargets {
		randIntn = func(int) int { return i }
		if got, want := lconnPicker(r), r.wTargets[i]; got != want {
			t.Fatalf("weight group at slot %d: got %s, want %s", i, got.URL, want.URL)
		}
	}
	r.Targets[0].End()
	r.Targets[0].End()
	r.Targets[1].End()

	zero := leastConnTable(t, "route add disabled / http://disabled/ weight 0")
	if got := zero.lookup("", "/", lconnPicker, prefixMatcher); got != nil {
		t.Fatalf("zero-weight single target must not be selected: %s", got.URL)
	}
	one := leastConnTable(t, "route add only / http://only/")
	if got := one.lookup("", "/", lconnPicker, prefixMatcher); got != one.route("", "/").Targets[0] {
		t.Fatalf("single target not selected: %v", got)
	}
}

func TestLeastConnConcurrent(t *testing.T) {
	r := &Route{Host: "", Path: "/"}
	for i := range 4 {
		r.addTarget("svc", mustParse(fmt.Sprintf("http://target-%d/", i)), -1, nil, nil)
	}
	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			for range 200 {
				target := lconnPicker(r)
				target.Begin()
				_ = target.Active()
				target.End()
			}
		})
	}
	wg.Wait()
	for _, target := range r.Targets {
		if got := target.Active(); got != 0 {
			t.Errorf("%s: active count %d, want 0", target.URL, got)
		}
	}
}

func TestLeastConnTableReload(t *testing.T) {
	previous := GetTable()
	defer SetTable(previous)
	lines := `route add svc / http://busy/
route add svc / http://idle/`
	first := leastConnTable(t, lines)
	SetTable(first)
	busy := first.route("", "/").Targets[0]
	busy.Begin()

	second := leastConnTable(t, lines)
	SetTable(second)
	if got := second.route("", "/").Targets[0].Active(); got != 1 {
		t.Fatalf("active work lost across reload: got %d, want 1", got)
	}
	if got := second.lookup("", "/", lconnPicker, prefixMatcher); got != second.route("", "/").Targets[1] {
		t.Fatalf("reloaded table did not avoid busy target: %v", got)
	}
	busy.End()
	if got := second.route("", "/").Targets[0].Active(); got != 0 {
		t.Fatalf("completion not visible after reload: got %d, want 0", got)
	}

	busy.Begin()
	third := leastConnTable(t, "route add svc / http://idle/")
	SetTable(third)
	if got := third.route("", "/").Targets[0].Active(); got != 0 {
		t.Fatalf("removed target count leaked to new route: %d", got)
	}
	busy.End()
}
