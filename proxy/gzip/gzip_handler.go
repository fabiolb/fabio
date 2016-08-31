// Copyright (c) 2016 Sebastian Mancke and eBay, both MIT licensed

// Package gzip provides an HTTP handler which compresses responses
// if the client supports this, the response is compressable and
// not already compressed.
//
// Based on https://github.com/smancke/handler/gzip
package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

const (
	headerVary            = "Vary"
	headerAcceptEncoding  = "Accept-Encoding"
	headerContentEncoding = "Content-Encoding"
	headerContentType     = "Content-Type"
	headerContentLength   = "Content-Length"
	encodingGzip          = "gzip"
)

var gzipWriterPool = sync.Pool{
	New: func() interface{} { return gzip.NewWriter(nil) },
}

// NewGzipHandler wraps an existing handler to transparently gzip the response
// body if the client supports it (via the Accept-Encoding header) and the
// response Content-Type matches the contentTypes expression.
func NewGzipHandler(h http.Handler, contentTypes *regexp.Regexp) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(headerVary, headerAcceptEncoding)

		if acceptsGzip(r) {
			gzWriter := NewGzipResponseWriter(w, contentTypes)
			defer gzWriter.Close()
			h.ServeHTTP(gzWriter, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

type GzipResponseWriter struct {
	writer       io.Writer
	gzipWriter   *gzip.Writer
	contentTypes *regexp.Regexp
	http.ResponseWriter
}

func NewGzipResponseWriter(w http.ResponseWriter, contentTypes *regexp.Regexp) *GzipResponseWriter {
	return &GzipResponseWriter{ResponseWriter: w, contentTypes: contentTypes}
}

func (grw *GzipResponseWriter) WriteHeader(code int) {
	if grw.writer == nil {
		if isCompressable(grw.Header(), grw.contentTypes) {
			grw.Header().Del(headerContentLength)
			grw.Header().Set(headerContentEncoding, encodingGzip)
			grw.gzipWriter = gzipWriterPool.Get().(*gzip.Writer)
			grw.gzipWriter.Reset(grw.ResponseWriter)

			grw.writer = grw.gzipWriter
		} else {
			grw.writer = grw.ResponseWriter
		}
	}
	grw.ResponseWriter.WriteHeader(code)
}

func (grw *GzipResponseWriter) Write(b []byte) (int, error) {
	if grw.writer == nil {
		if _, ok := grw.Header()[headerContentType]; !ok {
			// Set content-type if not present. Otherwise golang would make application/gzip out of that.
			grw.Header().Set(headerContentType, http.DetectContentType(b))
		}
		grw.WriteHeader(http.StatusOK)
	}
	return grw.writer.Write(b)
}

func (grw *GzipResponseWriter) Close() {
	if grw.gzipWriter != nil {
		grw.gzipWriter.Close()
		gzipWriterPool.Put(grw.gzipWriter)
	}
}

func isCompressable(header http.Header, contentTypes *regexp.Regexp) bool {
	// don't compress if it is already encoded
	if header.Get(headerContentEncoding) != "" {
		return false
	}
	return contentTypes.MatchString(header.Get(headerContentType))
}

func acceptsGzip(r *http.Request) bool {
	return strings.Contains(r.Header.Get(headerAcceptEncoding), encodingGzip)
}
