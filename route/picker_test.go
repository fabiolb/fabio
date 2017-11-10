package route

import (
	"net/url"
	"reflect"
	"testing"
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
