package consul

import (
	"reflect"
	"testing"

	"github.com/fabiolb/fabio/config"
)

func TestServiceRegistrationAddsProxyPortTagsOnlyToFabioService(t *testing.T) {
	cfg := &config.Consul{
		ServiceAddr:      "127.0.0.1:9998",
		ServiceName:      "fabio",
		ServiceTags:      []string{"custom-tag"},
		ServiceProxyTags: []string{"fabio-proxy-port=http:9999"},
		CheckScheme:      "http",
	}

	got, err := serviceRegistration(cfg, "fabio")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"custom-tag", "fabio-proxy-port=http:9999"}; !reflect.DeepEqual(got.Tags, want) {
		t.Fatalf("tags = %v, want %v", got.Tags, want)
	}

	alias, err := serviceRegistration(cfg, "backend")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"custom-tag"}; !reflect.DeepEqual(alias.Tags, want) {
		t.Fatalf("alias tags = %v, want %v", alias.Tags, want)
	}
}
