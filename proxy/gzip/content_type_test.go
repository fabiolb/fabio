package gzip

import "testing"

// TestContentTypes tests the content-type regexp that is used as
// an example in fabio.properties
func TestContentTypes(t *testing.T) {
	tests := []string{
		"text/foo",
		"text/foo; charset=UTF-8",
		"text/plain",
		"text/plain; charset=UTF-8",
		"application/json",
		"application/json; charset=UTF-8",
		"application/javascript",
		"application/javascript; charset=UTF-8",
		"application/font-woff",
		"application/font-woff; charset=UTF-8",
		"application/xml",
		"application/xml; charset=UTF-8",
		"vendor/vendor.foo+json",
		"vendor/vendor.foo+json; charset=UTF-8",
		"vendor/vendor.foo+xml",
		"vendor/vendor.foo+xml; charset=UTF-8",
	}

	for _, tt := range tests {
		tt := tt // capture loop var
		t.Run(tt, func(t *testing.T) {
			if !contentTypes.MatchString(tt) {
				t.Fatalf("%q does not match content types regexp", tt)
			}
		})
	}
}
