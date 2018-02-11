package metrics

import (
	"net/url"
	"os"
	"testing"
)

func TestParsePrefix(t *testing.T) {
	hostname = func() (string, error) { return "myhost", nil }
	os.Args = []string{"./myapp"}
	got, err := parsePrefix("{{clean .Hostname}}.{{clean .Exec}}")
	if err != nil {
		t.Fatalf("%v", err)
	}
	want := "myhost.myapp"
	if got != want {
		t.Errorf("ParsePrefix: got %v want %v", got, want)
	}

	got, err = parsePrefix("default")
	if err != nil {
		t.Fatalf("%v", err)
	}
	want = "myhost.myapp"
	if got != want {
		t.Errorf("ParsePrefix Old default style: got %v want %v", got, want)
	}
}

func TestTargetName(t *testing.T) {
	tests := []struct {
		service, host, path, target string
		name                        string
	}{
		{"s", "h", "p", "http://foo.com/bar", "s.h.p.foo_com"},
		{"s", "", "p", "http://foo.com/bar", "s._.p.foo_com"},
		{"s", "", "", "http://foo.com/bar", "s._._.foo_com"},
		{"", "", "", "http://foo.com/bar", "_._._.foo_com"},
		{"", "", "", "http://foo.com:1234/bar", "_._._.foo_com_1234"},
		{"", "", "", "http://1.2.3.4:1234/bar", "_._._.1_2_3_4_1234"},
	}

	for i, tt := range tests {
		u, err := url.Parse(tt.target)
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}

		got, err := TargetName(tt.service, tt.host, tt.path, u)
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}
		if want := tt.name; got != want {
			t.Errorf("%d: got %q want %q", i, got, want)
		}
	}
}
