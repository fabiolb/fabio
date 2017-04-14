// Copyright (c) 2016 Sebastian Mancke and eBay, both MIT licensed
package gzip

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"

	"github.com/fabiolb/fabio/assert"
)

var contentTypes = regexp.MustCompile(`^(text/.*|application/(javascript|json|font-woff|xml)|.*\+(json|xml))(;.*)?$`)

func Test_GzipHandler_CompressableType(t *testing.T) {
	server := httptest.NewServer(NewGzipHandler(test_text_handler(), contentTypes))

	assertEqual := assert.Equal(t)

	r, err := http.NewRequest("GET", server.URL, nil)
	assertEqual(err, nil)
	r.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(r)
	assertEqual(err, nil)

	assertEqual(resp.Header.Get("Content-Type"), "text/plain; charset=utf-8")
	assertEqual(resp.Header.Get("Content-Encoding"), "gzip")

	gzBytes, err := ioutil.ReadAll(resp.Body)
	assertEqual(err, nil)
	assertEqual(resp.Header.Get("Content-Length"), strconv.Itoa(len(gzBytes)))

	reader, err := gzip.NewReader(bytes.NewBuffer(gzBytes))
	assertEqual(err, nil)
	defer reader.Close()

	bytes, err := ioutil.ReadAll(reader)
	assertEqual(err, nil)

	assertEqual(string(bytes), "Hello World")
}

func Test_GzipHandler_NotCompressingTwice(t *testing.T) {
	server := httptest.NewServer(NewGzipHandler(test_already_compressed_handler(), contentTypes))

	assertEqual := assert.Equal(t)

	r, err := http.NewRequest("GET", server.URL, nil)
	assertEqual(err, nil)
	r.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(r)
	assertEqual(err, nil)

	assertEqual(resp.Header.Get("Content-Encoding"), "gzip")

	reader, err := gzip.NewReader(resp.Body)
	assertEqual(err, nil)
	defer reader.Close()

	bytes, err := ioutil.ReadAll(reader)
	assertEqual(err, nil)

	assertEqual(string(bytes), "Hello World")
}

func Test_GzipHandler_CompressableType_NoAccept(t *testing.T) {
	server := httptest.NewServer(NewGzipHandler(test_text_handler(), contentTypes))

	assertEqual := assert.Equal(t)

	r, err := http.NewRequest("GET", server.URL, nil)
	assertEqual(err, nil)
	r.Header.Set("Accept-Encoding", "none")

	resp, err := http.DefaultClient.Do(r)
	assertEqual(err, nil)

	assertEqual(resp.Header.Get("Content-Encoding"), "")

	bytes, err := ioutil.ReadAll(resp.Body)
	assertEqual(err, nil)

	assertEqual(string(bytes), "Hello World")
}

func Test_GzipHandler_NonCompressableType(t *testing.T) {
	server := httptest.NewServer(NewGzipHandler(test_binary_handler(), contentTypes))

	assertEqual := assert.Equal(t)

	r, err := http.NewRequest("GET", server.URL, nil)
	assertEqual(err, nil)
	r.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(r)
	assertEqual(err, nil)

	assertEqual(resp.Header.Get("Content-Encoding"), "")

	bytes, err := ioutil.ReadAll(resp.Body)
	assertEqual(err, nil)

	assertEqual(bytes, []byte{42})
}

func test_text_handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := []byte("Hello World")
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		w.Write(b)
	})
}

func test_binary_handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpg")
		w.Write([]byte{42})
	})
}

func test_already_compressed_handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gzWriter := gzip.NewWriter(w)
		gzWriter.Write([]byte("Hello World"))
		gzWriter.Close()
	})
}
