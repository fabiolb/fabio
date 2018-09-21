package route

import (
	"reflect"
	"sort"
	"testing"
)

func TestGlobCache(t *testing.T) {
	c := NewGlobCache(3)

	keys := func() []string {
		var kk []string
		for k := range c.m {
			kk = append(kk, k)
		}
		sort.Strings(kk)
		return kk
	}

	c.Get("a")
	if got, want := len(c.m), 1; got != want {
		t.Fatalf("got len %d want %d", got, want)
	}
	if got, want := keys(), []string{"a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := c.l, []string{"a", "", ""}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	c.Get("b")
	if got, want := len(c.m), 2; got != want {
		t.Fatalf("got len %d want %d", got, want)
	}
	if got, want := keys(), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := c.l, []string{"a", "b", ""}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	c.Get("c")
	if got, want := len(c.m), 3; got != want {
		t.Fatalf("got len %d want %d", got, want)
	}
	if got, want := keys(), []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := c.l, []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	c.Get("d")
	if got, want := len(c.m), 3; got != want {
		t.Fatalf("got len %d want %d", got, want)
	}
	if got, want := keys(), []string{"b", "c", "d"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := c.l, []string{"d", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
