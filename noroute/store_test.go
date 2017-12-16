package noroute

import (
	"testing"
)

func TestStoreSetGet(t *testing.T) {
	if got, want := GetHTML(), ""; got != want {
		t.Fatalf("got unset noroute html %q want %q", got, want)
	}

	SetHTML("foo")
	if got, want := GetHTML(), "foo"; got != want {
		t.Fatalf("got noroute html %q want %q", got, want)
	}
}
