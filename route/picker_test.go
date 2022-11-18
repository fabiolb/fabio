package route

import (
	"net/url"
	"reflect"
	"testing"
	"time"
)

var (
	fooDotCom = mustParse("http://foo.com/")
	barDotCom = mustParse("http://bar.com/")
)

func mustParse(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}

func TestRndPicker(t *testing.T) {
	r := &Route{Host: "www.bar.com", Path: "/foo"}
	r.addTarget("svc", fooDotCom, 0, nil, nil)
	r.addTarget("svc", barDotCom, 0, nil, nil)

	tests := []struct {
		rnd       int
		targetURL *url.URL
	}{
		{0, fooDotCom},
		{1, barDotCom},
	}

	prev := randIntn
	defer func() { randIntn = prev }()

	for i, tt := range tests {
		randIntn = func(int) int { return i }
		if got, want := rndPicker(r).URL, tt.targetURL; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}

func TestRRPicker(t *testing.T) {
	r := &Route{Host: "www.bar.com", Path: "/foo"}
	r.addTarget("svc", fooDotCom, 0, nil, nil)
	r.addTarget("svc", barDotCom, 0, nil, nil)

	tests := []*url.URL{fooDotCom, barDotCom, fooDotCom, barDotCom, fooDotCom, barDotCom}

	for i, tt := range tests {
		if got, want := rrPicker(r).URL, tt; !reflect.DeepEqual(got, want) {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}

// This is an improved version of the previous UnixNano implementation
// This one does not overflow on 32 bit platforms, it casts to int after
// doing mod.  doing it before caused overflows.
var oldRandInt = func(n int) int {
	if n == 0 {
		return 0
	}
	return int(time.Now().UnixNano() / int64(time.Microsecond) % int64(n))
}

var result int // prevent compiler optimization
func BenchmarkOldRandIntn(b *testing.B) {
	var r int // more shields against compiler optimization
	for i := 0; i < b.N; i++ {
		r = oldRandInt(i)
	}
	result = r
}
func BenchmarkMathRandIntn(b *testing.B) {
	var r int // more shields against compiler optimization
	for i := 0; i < b.N; i++ {
		r = randIntn(i)
	}
	result = r
}
