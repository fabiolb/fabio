package auth

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fabiolb/fabio/config"
)

func TestExternal_Authorized(t *testing.T) {
	var authRequest *http.Request
	var authBody []byte
	var authBodyErr error
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authRequest = r.Clone(r.Context())
		authBody, authBodyErr = io.ReadAll(r.Body)

		w.Header().Set("X-Auth-Request-User", "alice")
		w.Header().Add("X-Auth-Request-Groups", "staff")
		w.Header().Set("X-Unconfigured", "from-auth")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("accepted"))
	}))
	defer authServer.Close()

	externalAuth, err := newExternalAuth(config.ExternalAuth{
		Endpoint:          authServer.URL + "/oauth2/auth?fixed=1",
		SetAuthHeaders:    []string{"X-Auth-Request-User", "X-Auth-Missing"},
		AppendAuthHeaders: []string{"X-Auth-Request-Groups"},
	}, noRedirectClient())
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "http://app.example/private/resource?x=1", strings.NewReader("payload"))
	request.Host = "app.example"
	request.RemoteAddr = "192.0.2.10:4321"
	request.RequestURI = "/private//resource?x=1"
	request.Header.Set("X-Existing", "preserved")
	request.Header.Set("X-Auth-Request-User", "spoofed")
	request.Header.Set("X-Auth-Missing", "spoofed")
	request.Header.Set("X-Auth-Request-Groups", "existing")
	request.Header.Set("X-Unconfigured", "original")

	decision := externalAuth.Authorized(request, httptest.NewRecorder())
	if got, want := decision, (AuthDecision{Authorized: true}); got != want {
		t.Fatalf("decision = %#v, want %#v", got, want)
	}

	if authRequest == nil {
		t.Fatal("auth endpoint was not called")
	}
	if authBodyErr != nil {
		t.Fatalf("read auth request body: %s", authBodyErr)
	}
	if len(authBody) != 0 {
		t.Fatalf("auth request body = %q, want empty", authBody)
	}
	if got, want := authRequest.Method, http.MethodGet; got != want {
		t.Errorf("auth method = %q, want %q", got, want)
	}
	if got, want := authRequest.URL.RequestURI(), "/oauth2/auth?fixed=1"; got != want {
		t.Errorf("auth URI = %q, want %q", got, want)
	}
	if got, want := authRequest.Header.Get("X-Forwarded-Method"), http.MethodPost; got != want {
		t.Errorf("X-Forwarded-Method = %q, want %q", got, want)
	}
	if got, want := authRequest.Header.Get("X-Forwarded-Uri"), "/private//resource?x=1"; got != want {
		t.Errorf("X-Forwarded-Uri = %q, want %q", got, want)
	}
	if got, want := authRequest.Header.Get("X-Forwarded-Host"), "app.example"; got != want {
		t.Errorf("X-Forwarded-Host = %q, want %q", got, want)
	}
	if got, want := authRequest.Header.Get("X-Forwarded-Proto"), "http"; got != want {
		t.Errorf("X-Forwarded-Proto = %q, want %q", got, want)
	}
	if got, want := authRequest.Header.Get("X-Forwarded-For"), "192.0.2.10"; got != want {
		t.Errorf("X-Forwarded-For = %q, want %q", got, want)
	}
	if got, want := authRequest.Header.Get("Forwarded"), "for=192.0.2.10; proto=http"; got != want {
		t.Errorf("Forwarded = %q, want %q", got, want)
	}
	if got, want := authRequest.Header.Get("X-Existing"), "preserved"; got != want {
		t.Errorf("cloned request header = %q, want %q", got, want)
	}

	if got, want := request.Header.Get("X-Auth-Request-User"), "alice"; got != want {
		t.Errorf("set auth header = %q, want %q", got, want)
	}
	if got := request.Header.Get("X-Auth-Missing"); got != "" {
		t.Errorf("missing set auth header = %q, want empty", got)
	}
	if got, want := request.Header.Values("X-Auth-Request-Groups"), []string{"existing", "staff"}; !stringSlicesEqual(got, want) {
		t.Errorf("appended auth headers = %#v, want %#v", got, want)
	}
	if got, want := request.Header.Get("X-Unconfigured"), "original"; got != want {
		t.Errorf("unconfigured auth response header changed request: got %q want %q", got, want)
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(body), "payload"; got != want {
		t.Errorf("original request body = %q, want %q", got, want)
	}
}

func TestExternal_DeniedResponse(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{name: "unauthorized", status: http.StatusUnauthorized},
		{name: "forbidden", status: http.StatusForbidden},
		{name: "redirect", status: http.StatusFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", "/oauth2/sign_in")
				w.Header().Set("WWW-Authenticate", `Bearer realm="test"`)
				w.Header().Add("Set-Cookie", "session=abc; Path=/")
				w.Header().Set("Connection", "X-Hop")
				w.Header().Set("X-Hop", "do-not-forward")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte("denied"))
			}))
			defer authServer.Close()

			externalAuth, err := newExternalAuth(config.ExternalAuth{Endpoint: authServer.URL + "/oauth2/auth"}, noRedirectClient())
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			decision := externalAuth.Authorized(httptest.NewRequest(http.MethodGet, "http://app.example/private", nil), recorder)
			if got, want := decision, (AuthDecision{Done: true}); got != want {
				t.Fatalf("decision = %#v, want %#v", got, want)
			}
			if got, want := recorder.Code, tt.status; got != want {
				t.Errorf("status = %d, want %d", got, want)
			}
			if got, want := recorder.Body.String(), "denied"; got != want {
				t.Errorf("body = %q, want %q", got, want)
			}
			if got, want := recorder.Header().Get("Location"), "/oauth2/sign_in"; got != want {
				t.Errorf("Location = %q, want %q", got, want)
			}
			if got, want := recorder.Header().Get("WWW-Authenticate"), `Bearer realm="test"`; got != want {
				t.Errorf("WWW-Authenticate = %q, want %q", got, want)
			}
			if got, want := recorder.Header().Values("Set-Cookie"), []string{"session=abc; Path=/"}; !stringSlicesEqual(got, want) {
				t.Errorf("Set-Cookie = %#v, want %#v", got, want)
			}
			if got := recorder.Header().Get("Connection"); got != "" {
				t.Errorf("Connection header forwarded: %q", got)
			}
			if got := recorder.Header().Get("X-Hop"); got != "" {
				t.Errorf("Connection-nominated header forwarded: %q", got)
			}
		})
	}
}

func TestExternal_TransportFailureFailsClosed(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("auth unavailable")
	})}
	externalAuth, err := newExternalAuth(config.ExternalAuth{Endpoint: "http://auth.example/oauth2/auth"}, client)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	decision := externalAuth.Authorized(httptest.NewRequest(http.MethodGet, "http://app.example/private", nil), recorder)
	if got, want := decision, (AuthDecision{Done: true}); got != want {
		t.Fatalf("decision = %#v, want %#v", got, want)
	}
	if got, want := recorder.Code, http.StatusBadGateway; got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
}

func TestExternal_TimeoutFailsClosed(t *testing.T) {
	client := &http.Client{
		Timeout: 10 * time.Millisecond,
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			<-r.Context().Done()
			return nil, r.Context().Err()
		}),
	}
	externalAuth, err := newExternalAuth(config.ExternalAuth{Endpoint: "http://auth.example/oauth2/auth"}, client)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	decision := externalAuth.Authorized(httptest.NewRequest(http.MethodGet, "http://app.example/private", nil), recorder)
	if got, want := decision, (AuthDecision{Done: true}); got != want {
		t.Fatalf("decision = %#v, want %#v", got, want)
	}
	if got, want := recorder.Code, http.StatusBadGateway; got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
}

func TestLoadAuthSchemes_ExternalRequiresClient(t *testing.T) {
	_, err := LoadAuthSchemes(map[string]config.AuthScheme{
		"oauth": {
			Name:     "oauth",
			Type:     "external",
			External: config.ExternalAuth{Endpoint: "http://auth.example/oauth2/auth"},
		},
	}, nil)
	if err == nil || err.Error() != "external auth requires an HTTP client" {
		t.Fatalf("error = %v, want external auth client error", err)
	}
}

func noRedirectClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
