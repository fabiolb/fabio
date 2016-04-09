package cert

import "testing"

func TestBase(t *testing.T) {
	tests := []struct {
		in, out, err string
	}{
		{"", "", ""},
		{"http://foo.com/x/y", "http://foo.com/x", ""},
		{"http://foo.com/x/y?p=q", "http://foo.com/x?p=q", ""},
	}

	for i, tt := range tests {
		u, err := base(tt.in)
		if err != nil {
			if got, want := err.Error(), tt.err; got != want {
				t.Errorf("%d: got %v want %v", i, got, want)
				continue
			}
		}
		if tt.err != "" {
			t.Errorf("%d: got nil want %v", i, tt.err)
			continue
		}
		if got, want := u, tt.out; got != want {
			t.Errorf("%d: got %v want %v", i, got, want)
		}
	}
}

func TestReplaceSuffix(t *testing.T) {
	if got, want := replaceSuffix("ab", "b", "c"), "ac"; got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
