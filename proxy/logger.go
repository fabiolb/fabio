package proxy

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	pattern *Pattern
)

type Pattern struct {
	converts []Converter
}

type Converter interface {
	convert(t time.Time, r http.Request) string
}

type RequestConvert struct {
}

func (RequestConvert) convert(t time.Time, r http.Request) string {
	return "\"" + r.Method + " " + r.URL.Path + " " + r.Proto + "\""
}

type RemoteAddrConvert struct {
}

func (RemoteAddrConvert) convert(t time.Time, r http.Request) string {
	return r.RemoteAddr[:strings.Index(r.RemoteAddr, ":")]
}

type BodyBytesSentConvert struct {
}

func (BodyBytesSentConvert) convert(t time.Time, r http.Request) string {
	return fmt.Sprintf("%d", uint64(r.ContentLength))
}

type HttpRefererConvert struct {
}

func (HttpRefererConvert) convert(t time.Time, r http.Request) string {
	return r.Referer()
}

type HttpUserAgentConvert struct {
}

func (HttpUserAgentConvert) convert(t time.Time, r http.Request) string {
	return r.UserAgent()
}

type ServerNameConvert struct {
}

func (ServerNameConvert) convert(t time.Time, r http.Request) string {
	return r.Host
}

type UpstreamAddrConvert struct {
}

func (UpstreamAddrConvert) convert(t time.Time, r http.Request) string {
	return r.URL.Host
}

type RequestArgsConvert struct {
}

func (RequestArgsConvert) convert(t time.Time, r http.Request) string {
	return r.URL.RawQuery
}

type UpstreamResponseTimeConvert struct {
}

func (UpstreamResponseTimeConvert) convert(t time.Time, r http.Request) string {
	return fmt.Sprintf("%.4f", time.Since(t).Seconds())
}

type TimeConvert struct {
}

func (TimeConvert) convert(t time.Time, r http.Request) string {
	return t.Format(time.RFC3339)
}

type DefaultConvert struct {
}

func (DefaultConvert) convert(t time.Time, r http.Request) string {
	return ""
}

func patternParse(pattern string) *Pattern {
	flags := strings.Split(pattern, " ")

	c := make([]Converter, len(flags))

	for i, flag := range flags {
		switch flag {
		case "$remote_addr":
			c[i] = &RemoteAddrConvert{}
		case "$time":
			c[i] = &TimeConvert{}
		case "$request":
			c[i] = &RequestConvert{}
		case "$body_bytes_sent":
			c[i] = &BodyBytesSentConvert{}
		case "$http_referer":
			c[i] = &HttpRefererConvert{}
		case "$http_user_agent":
			c[i] = &HttpUserAgentConvert{}
		case "$server_name":
			c[i] = &ServerNameConvert{}
		case "$upstream_addr":
			c[i] = &UpstreamAddrConvert{}
		case "$upstream_response_time":
			c[i] = &UpstreamResponseTimeConvert{}
		case "$request_args":
			c[i] = &RequestArgsConvert{}
		default:
			c[i] = &DefaultConvert{}
		}
	}

	return &Pattern{converts: c}
}

func logger(log *log.Logger, format string, t time.Time, r *http.Request) {

	if pattern == nil {
		pattern = patternParse(format)
	}

	str := ""
	for _, c := range pattern.converts {
		str += c.convert(t, *r) + " "
	}

	log.Println(str)
}
