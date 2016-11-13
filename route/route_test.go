package route

import (
	"net/url"
	"reflect"
	"testing"
)

func mustParse(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}

func TestNewRoute(t *testing.T) {
	r := &Route{Host: "www.bar.com", Path: "/foo"}
	if got, want := r.Path, "/foo"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestAddTarget(t *testing.T) {
	u := mustParse("http://foo.com/")

	r := &Route{Host: "www.bar.com", Path: "/foo"}
	r.addTarget("service", u, 0, nil)

	if got, want := len(r.Targets), 1; got != want {
		t.Errorf("target length: got %d want %d", got, want)
	}
	if got, want := r.Targets[0].URL, u; got != want {
		t.Errorf("target url: got %s want %s", got, want)
	}
	config := []string{"route add service www.bar.com/foo http://foo.com/"}
	if got, want := r.config(false), config; !reflect.DeepEqual(got, want) {
		t.Errorf("config: got %q want %q", got, want)
	}
}

func TestDelService(t *testing.T) {
	u1, u2 := mustParse("http://foo.com/"), mustParse("http://bar.com/")

	r := &Route{Host: "www.bar.com", Path: "/foo"}
	r.addTarget("serviceA", u1, 0, nil)
	r.addTarget("serviceB", u2, 0, nil)
	r.delService("serviceA")

	config := []string{"route add serviceB www.bar.com/foo http://bar.com/"}
	if got, want := r.config(false), config; !reflect.DeepEqual(got, want) {
		t.Errorf("config: got %q want %q", got, want)
	}
}
