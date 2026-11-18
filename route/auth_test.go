package route

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/fabiolb/fabio/auth"
)

type testAuth struct {
	decision auth.AuthDecision
}

func (t *testAuth) Authorized(r *http.Request, w http.ResponseWriter) auth.AuthDecision {
	return t.decision
}

type responseWriter struct {
	header  http.Header
	code    int
	written []byte
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.written = append(rw.written, b...)
	return len(rw.written), nil
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.code = statusCode
}

func TestTarget_Authorized(t *testing.T) {
	tests := []struct {
		name        string
		authScheme  string
		authSchemes map[string]auth.AuthScheme
		out         auth.AuthDecision
	}{
		{
			name:       "matches correct auth scheme",
			authScheme: "mybasic",
			authSchemes: map[string]auth.AuthScheme{
				"mybasic": &testAuth{decision: auth.AuthDecision{Authorized: true}},
			},
			out: auth.AuthDecision{Authorized: true},
		},
		{
			name:       "returns true when scheme is empty",
			authScheme: "",
			authSchemes: map[string]auth.AuthScheme{
				"mybasic": &testAuth{},
			},
			out: auth.AuthDecision{Authorized: true},
		},
		{
			name:       "returns false when scheme is unknown",
			authScheme: "foobar",
			authSchemes: map[string]auth.AuthScheme{
				"mybasic": &testAuth{decision: auth.AuthDecision{Authorized: true}},
			},
			out: auth.AuthDecision{},
		},
		{
			name:       "returns completed response decision from scheme",
			authScheme: "external",
			authSchemes: map[string]auth.AuthScheme{
				"external": &testAuth{decision: auth.AuthDecision{Done: true}},
			},
			out: auth.AuthDecision{Done: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := &Target{
				AuthScheme: tt.authScheme,
			}

			if got, want := target.Authorized(&http.Request{}, &responseWriter{}, tt.authSchemes), tt.out; !reflect.DeepEqual(got, want) {
				t.Errorf("got %v want %v", got, want)
			}
		})
	}
}
