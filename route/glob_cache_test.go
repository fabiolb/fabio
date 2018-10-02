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
		c.m.Range(func(k, v interface{}) bool {
			kk = append(kk, k.(string))
			return true
		})
		sort.Strings(kk)
		return kk
	}

	c.Get("a")
	// TODO  add back in when sync.Map supports Len function
	// TODO https://github.com/golang/go/issues/20680
	//if got, want := len(c.m), 1; got != want {
	//	t.Fatalf("got len %d want %d", got, want)
	//}
	if got, want := keys(), []string{"a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := c.l, []string{"a", "", ""}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	c.Get("b")
	// TODO  add back in when sync.Map supports Len function
	// TODO https://github.com/golang/go/issues/20680
	//if got, want := len(c.m), 2; got != want {
	//	t.Fatalf("got len %d want %d", got, want)
	//}
	if got, want := keys(), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := c.l, []string{"a", "b", ""}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	c.Get("c")
	// TODO  add back in when sync.Map supports Len function
	// TODO https://github.com/golang/go/issues/20680
	//if got, want := len(c.m), 3; got != want {
	//	t.Fatalf("got len %d want %d", got, want)
	//}
	if got, want := keys(), []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := c.l, []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	c.Get("d")
	// TODO  add back in when sync.Map supports Len function
	// TODO https://github.com/golang/go/issues/20680
	//if got, want := len(c.m), 3; got != want {
	//	t.Fatalf("got len %d want %d", got, want)
	//}
	if got, want := keys(), []string{"b", "c", "d"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if got, want := c.l, []string{"d", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
