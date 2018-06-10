package fastcgi

import (
	"io"
	"net/http"
	"time"
)

type staticFcgiBackend struct {
	OptionsFunc        func(map[string]string) (*http.Response, error)
	HeadFunc           func(map[string]string) (*http.Response, error)
	GetFunc            func(map[string]string) (*http.Response, error)
	PostFunc           func(map[string]string, string, string, io.Reader, int64) (*http.Response, error)
	SetReadTimeoutFunc func(time.Duration) error
	SetSendTimeoutFunc func(time.Duration) error
	StderrFunc         func() string
	CloseFunc          func()
}

func (b *staticFcgiBackend) Options(params map[string]string) (*http.Response, error) {
	return b.OptionsFunc(params)
}

func (b *staticFcgiBackend) Head(params map[string]string) (*http.Response, error) {
	return b.HeadFunc(params)
}

func (b *staticFcgiBackend) Get(params map[string]string) (*http.Response, error) {
	return b.GetFunc(params)
}

func (b *staticFcgiBackend) SetReadTimeout(dur time.Duration) error {
	return b.SetReadTimeoutFunc(dur)
}

func (b *staticFcgiBackend) SetSendTimeout(dur time.Duration) error {
	return b.SetSendTimeoutFunc(dur)
}

func (b *staticFcgiBackend) Stderr() string {
	return b.StderrFunc()
}

func (b *staticFcgiBackend) Post(params map[string]string, method string, bodyType string, body io.Reader, l int64) (*http.Response, error) {
	return b.PostFunc(params, method, bodyType, body, l)
}

func (b *staticFcgiBackend) Close() {}
