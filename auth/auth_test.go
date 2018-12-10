package auth

import (
	"testing"

	"github.com/fabiolb/fabio/config"
)

func TestLoadAuthSchemes(t *testing.T) {

	t.Run("should fail when auth scheme fails to load", func(t *testing.T) {
		_, err := LoadAuthSchemes(map[string]config.AuthScheme{
			"myauth": {
				Name: "myauth",
				Type: "basic",
				Basic: config.BasicAuth{
					File: "/some/non/existent/file",
				},
			},
		})

		const errorText = "open /some/non/existent/file: no such file or directory"

		if err.Error() != errorText {
			t.Fatalf("got %s, want %s", err.Error(), errorText)
		}
	})

	t.Run("should return an error when auth type is unknown", func(t *testing.T) {
		_, err := LoadAuthSchemes(map[string]config.AuthScheme{
			"myauth": {
				Name: "myauth",
				Type: "foo",
			},
		})

		const errorText = "unknown auth type 'foo'"

		if err.Error() != errorText {
			t.Fatalf("got %s, want %s", err.Error(), errorText)
		}
	})

	t.Run("should load multiple auth schemes", func(t *testing.T) {
		myauth, err := createBasicAuthFile("foo:bar")
		if err != nil {
			t.Fatalf("could not create file on disk %s", err)
		}

		myotherauth, err := createBasicAuthFile("bar:foo")
		if err != nil {
			t.Fatalf("could not create file on disk %s", err)
		}

		result, err := LoadAuthSchemes(map[string]config.AuthScheme{
			"myauth": {
				Name: "myauth",
				Type: "basic",
				Basic: config.BasicAuth{
					File: myauth,
				},
			},
			"myotherauth": {
				Name: "myotherauth",
				Type: "basic",
				Basic: config.BasicAuth{
					File: myotherauth,
				},
			},
		})

		if len(result) != 2 {
			t.Fatalf("expected 2 auth schemes, got %d", len(result))
		}
	})
}
