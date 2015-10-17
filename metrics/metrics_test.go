package metrics

import (
	"net/url"
	"os"
	"testing"
)

func TestDefaultPrefix(t *testing.T) {
	hostname = func() (string, error) { return "myhost", nil }
	os.Args = []string{"./myapp"}
	if got, want := defaultPrefix(), "myhost.myapp"; got != want {
		t.Errorf("got %v want %v", got, want)
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
		if got, want := TargetName(tt.service, tt.host, tt.path, u), tt.name; got != want {
			t.Errorf("%d: got %q want %q", i, got, want)
		}
	}
}
