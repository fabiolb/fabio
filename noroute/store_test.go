package noroute

import (
    "testing"
)

func TestStoreSetGet(t *testing.T) {
    got := GetHTML()
    if got != "" {
        t.Fatalf("Expected unset noroute html to be an empty string, got %s", got)
    }

    want := "<blink>Fancy!</blink>"
    SetHTML(want)
    got = GetHTML()
    if got != want {
        t.Fatalf("got %s, want %s", got, want)
    }
}
