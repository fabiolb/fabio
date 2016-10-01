package cert

import (
	"reflect"
	"testing"

	"github.com/hashicorp/consul/api"
)

func TestParseConsulURL(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		config *api.Config
		key    string
		errstr string
	}{
		{
			name:   "empty url",
			errstr: "invalid url",
		},
		{
			name:   "invalid url",
			in:     "this is not a url",
			errstr: "invalid url",
		},
		{
			name:   "no kv store url",
			in:     "http://localhost:8500/path/to/cert",
			errstr: "missing prefix: /v1/kv/",
		},
		{
			name:   "url without token",
			in:     "http://localhost:8500/v1/kv/path/to/cert",
			config: &api.Config{Address: "localhost:8500", Scheme: "http"},
			key:    "path/to/cert",
		},
		{
			name:   "https url",
			in:     "https://localhost:8500/v1/kv/path/to/cert",
			config: &api.Config{Address: "localhost:8500", Scheme: "https"},
			key:    "path/to/cert",
		},
		{
			name:   "url with token",
			in:     "http://localhost:8500/v1/kv/path/to/cert?token=123",
			config: &api.Config{Address: "localhost:8500", Scheme: "http", Token: "123"},
			key:    "path/to/cert",
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop var
		t.Run(tt.name, func(t *testing.T) {
			config, key, err := parseConsulURL(tt.in)
			var errstr string
			if err != nil {
				errstr = err.Error()
			}
			if got, want := errstr, tt.errstr; got != want {
				t.Fatalf("got err %q want %q", got, want)
			}
			if errstr != "" || tt.errstr != "" {
				return
			}
			if got, want := key, tt.key; got != want {
				t.Errorf("got key %q want %q", got, want)
			}
			if got, want := config, tt.config; !reflect.DeepEqual(got, want) {
				t.Errorf("got config %+v want %+v", got, want)
			}
		})
	}
}
