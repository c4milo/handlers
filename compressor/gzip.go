// Package compressor implements GZIP compression. This is highly based on
// https://github.com/phyber/negroni-gzip, with minor changes.
package compressor

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/c4milo/handlers/internal"
)

const (
	gzipEncoding    = "gzip"
	deflateEncoding = "deflate"

	acceptEncoding  = "Accept-Encoding"
	contentEncoding = "Content-Encoding"
	contentLength   = "Content-Length"
	contentType     = "Content-Type"
	vary            = "Vary"
	secWebSocketKey = "Sec-WebSocket-Key"

	BestCompression    = gzip.BestCompression
	BestSpeed          = gzip.BestSpeed
	DefaultCompression = gzip.DefaultCompression
	NoCompression      = gzip.NoCompression
)

// http://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
type option func(*handler)

// Internal handler
type handler struct {
	compressionLevel int
}

// GzipLevel allows configuring GZIP compression level.
// Options:
// * compressor.BestCompression
// * compressor.BestSpeed
// * compressor.DefaultCompression
// * compressor.NoCompression
func GzipLevel(l int) option {
	return func(h *handler) {
		h.compressionLevel = l
	}
}

// GZIP response writer wrapper
type responseWriter struct {
	internal.ResponseWriter
	gzipWriter *gzip.Writer
}

// Write writes bytes to the gzip.Writer. It will also set the Content-Type
// header using the net/http library content type detection if the Content-Type
// header was not set yet.
func (rw responseWriter) Write(b []byte) (int, error) {
	if !rw.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet
		rw.WriteHeader(http.StatusOK)
	}

	if rw.Header().Get(contentType) == "" {
		rw.Header().Set(contentType, http.DetectContentType(b))
	}

	size, err := rw.gzipWriter.Write(b)
	return size, err
}

// GzipHandler applies GZIP compression to the response body, except in the following
// scenarios:
// * The response body is already compressed using gzip or deflate
// * The request's Accept-Encoding header does not announce gzip support
// * The request is upgrading to a websocket connection.
func GzipHandler(h http.Handler, opts ...option) http.Handler {
	// Default options
	handler := &handler{
		compressionLevel: gzip.DefaultCompression,
	}

	for _, opt := range opts {
		opt(handler)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header
		if !strings.Contains(hdr.Get(acceptEncoding), gzipEncoding) {
			h.ServeHTTP(w, r)
			return
		}

		// Skip compression if body comes compressed already.
		curEncoding := hdr.Get(contentEncoding)
		if curEncoding == gzipEncoding ||
			curEncoding == deflateEncoding {
			h.ServeHTTP(w, r)
			return
		}

		// This handler does not support websockets compression
		if hdr.Get(secWebSocketKey) != "" {
			h.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, handler.compressionLevel)
		if err != nil {
			h.ServeHTTP(w, r)
			return
		}
		defer gz.Close()

		headers := w.Header()
		headers.Set(contentEncoding, gzipEncoding)
		headers.Set(vary, acceptEncoding)

		rw := responseWriter{internal.NewResponseWriter(w), gz}
		h.ServeHTTP(rw, r)
	})
}
