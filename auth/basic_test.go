package auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/fabiolb/fabio/config"
	"github.com/fabiolb/fabio/uuid"
)

type responseWriter struct {
	header  http.Header
	code    int
	written []byte
}

func (rw *responseWriter) Header() http.Header {
	if rw.header == nil {
		rw.header = map[string][]string{}
	}
	return rw.header
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.written = append(rw.written, b...)
	return len(rw.written), nil
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.code = statusCode
}

func createBasicAuthFile(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()

	filename := fmt.Sprintf("%s/%s", dir, uuid.NewUUID())

	err := os.WriteFile(filename, []byte(contents), 0666)
	if err != nil {
		t.Fatalf("could not write basic auth password file: %s", err)
	}

	return filename
}

func createBasicAuth(t *testing.T, user string, password string) AuthScheme {
	t.Helper()
	contents := fmt.Sprintf("%s:%s", user, password)

	filename := createBasicAuthFile(t, contents)

	basicAuth, err := newBasicAuth(config.BasicAuth{
		File:  filename,
		Realm: "testrealm",
	})
	if err != nil {
		t.Fatalf("could not create basic auth: %s", err)
	}

	return basicAuth
}

func basicAuthHeader(username, password string) []string {
	auth := []byte(username + ":" + password)
	return []string{"Basic " + base64.StdEncoding.EncodeToString(auth)}
}

func TestNewBasicAuth(t *testing.T) {

	t.Run("should create a basic auth scheme from the supplied config", func(t *testing.T) {
		filename := createBasicAuthFile(t, "foo:bar")

		_, err := newBasicAuth(config.BasicAuth{
			File: filename,
		})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("should log a warning when credentials are malformed", func(t *testing.T) {
		old := log.Writer()
		defer log.SetOutput(old)
		var buf bytes.Buffer
		log.SetOutput(&buf)

		filename := createBasicAuthFile(t, "foosdlijdgohdgdbar")

		_, err := newBasicAuth(config.BasicAuth{
			File: filename,
		})
		if err != nil {
			t.Fatal(err)
		}
		want := "[WARN] Error processing htpasswd file: malformed line, no colon: foosdlijdgohdgdbar"
		if got := buf.String(); !strings.Contains(got, want) {
			t.Fatalf("log does not contain expected:\nlog: %q\nexpected: %q", got, want)
		}
	})
}

func TestBasic_Authorised(t *testing.T) {
	basicAuth := createBasicAuth(t, "foo", "bar")

	tests := []struct {
		name string
		req  *http.Request
		res  http.ResponseWriter
		want bool
	}{
		{
			"correct credentials should be authorized",
			&http.Request{
				Header: http.Header{
					"Authorization": basicAuthHeader("foo", "bar"),
				},
			},
			&responseWriter{},
			true,
		},
		{
			"incorrect credentials should not be authorized",
			&http.Request{
				Header: http.Header{
					"Authorization": basicAuthHeader("baz", "blarg"),
				},
			},
			&responseWriter{},
			false,
		},
		{
			"missing Authorization header should not be authorized",
			&http.Request{
				Header: http.Header{},
			},
			&responseWriter{},
			false,
		},
		{
			"malformed Authorization header should not be authorized",
			&http.Request{
				Header: http.Header{
					"Authorization": []string{"malformed"},
				},
			},
			&responseWriter{},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := basicAuth.Authorized(tt.req, tt.res); got != tt.want {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestBasic_Authorised_should_fail_without_htpasswd_file(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		filename := createBasicAuthFile(t, "foo:bar")
		// We create a basic auth with periodic refresh.
		basicAuth, err := newBasicAuth(config.BasicAuth{
			File:    filename,
			Refresh: time.Second,
		})
		if err != nil {
			t.Error(err)
		}

		r := &http.Request{
			Header: http.Header{
				"Authorization": basicAuthHeader("foo", "bar"),
			},
		}
		w := &responseWriter{}

		// We want the first call to Authorized() to succeed.
		if got, want := basicAuth.Authorized(r, w), true; got != want {
			t.Errorf("got %v want %v", got, want)
		}

		// Since refresh is enabled, removing the htpasswd file will cause the next
		// call to Authorized() to fail, as documented.
		if err := os.Remove(filename); err != nil {
			t.Fatalf("removing htpasswd file: %s", err)
		}

		// Wait to ensure that the htpasswd file refresh happened.
		// This happens within a syntest.Test(), so it will take zero time.
		time.Sleep(2 * time.Second)
		// Before exiting from the synctest bubble, we must terminate the basicAuth
		// goroutine.
		// We don't close the channel because closing is async. The send instead, since
		// the channel is unbuffered, ensures that the test waits for the goroutine
		// to receive and terminate.
		basicAuth.done <- struct{}{}
		synctest.Wait()
		// We want the second call to Authorized() to fail.
		if got, want := basicAuth.Authorized(r, w), false; got != want {
			t.Errorf("got %v want %v", got, want)
		}
	})
}

func TestBasic_Authorized_should_set_www_realm_header(t *testing.T) {
	basicAuth := createBasicAuth(t, "foo", "bar")

	rw := &responseWriter{}

	_ = basicAuth.Authorized(&http.Request{Header: http.Header{}}, rw)

	got := rw.Header().Get("WWW-Authenticate")
	want := `Basic realm="testrealm"`

	if got != want {
		t.Errorf("got '%s', want '%s'", got, want)
	}
}
