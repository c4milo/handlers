// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package internal

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hooklift/assert"
)

type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func newCloseNotifyingRecorder() *closeNotifyingRecorder {
	return &closeNotifyingRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

func (c *closeNotifyingRecorder) close() {
	c.closed <- true
}

func (c *closeNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}

type hijackableResponse struct {
	Hijacked bool
}

func newHijackableResponse() *hijackableResponse {
	return &hijackableResponse{}
}

func (h *hijackableResponse) Header() http.Header           { return nil }
func (h *hijackableResponse) Write(buf []byte) (int, error) { return 0, nil }
func (h *hijackableResponse) WriteHeader(code int)          {}
func (h *hijackableResponse) Flush()                        {}
func (h *hijackableResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.Hijacked = true
	return nil, nil, nil
}

func TestResponseWriterWritingString(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	rw.Write([]byte("Hello world"))

	assert.Equals(t, rw.Status(), rec.Code)
	assert.Equals(t, rec.Body.String(), "Hello world")
	assert.Equals(t, rw.Status(), http.StatusOK)
	assert.Equals(t, rw.Size(), 11)
	assert.Equals(t, rw.Written(), true)
}

func TestResponseWriterWritingStrings(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	rw.Write([]byte("Hello world"))
	rw.Write([]byte("foo bar bat baz"))

	assert.Equals(t, rec.Code, rw.Status())
	assert.Equals(t, "Hello worldfoo bar bat baz", rec.Body.String())
	assert.Equals(t, http.StatusOK, rw.Status())
	assert.Equals(t, 26, rw.Size())
}

func TestResponseWriterWritingHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	rw.WriteHeader(http.StatusNotFound)

	assert.Equals(t, rw.Status(), rec.Code)
	assert.Equals(t, "", rec.Body.String())
	assert.Equals(t, http.StatusNotFound, rw.Status())
	assert.Equals(t, 0, rw.Size())
}

func TestResponseWriterBefore(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)
	result := ""

	rw.Before(func(ResponseWriter) {
		result += "foo"
	})
	rw.Before(func(ResponseWriter) {
		result += "bar"
	})

	rw.WriteHeader(http.StatusNotFound)

	assert.Equals(t, rw.Status(), rec.Code)
	assert.Equals(t, "", rec.Body.String())
	assert.Equals(t, http.StatusNotFound, rw.Status())
	assert.Equals(t, 0, rw.Size())
	assert.Equals(t, "barfoo", result)
}

func TestResponseWriterHijack(t *testing.T) {
	hijackable := newHijackableResponse()
	rw := NewResponseWriter(hijackable)
	hijacker, ok := rw.(http.Hijacker)
	assert.Equals(t, true, ok)
	_, _, err := hijacker.Hijack()
	if err != nil {
		t.Error(err)
	}
	assert.Equals(t, true, hijackable.Hijacked)
}

func TestResponseWriteHijackNotOK(t *testing.T) {
	hijackable := new(http.ResponseWriter)
	rw := NewResponseWriter(*hijackable)
	hijacker, ok := rw.(http.Hijacker)
	assert.Equals(t, true, ok)
	_, _, err := hijacker.Hijack()

	assert.Equals(t, "the ResponseWriter doesn't support the Hijacker interface", err.Error())
}

func TestResponseWriterCloseNotify(t *testing.T) {
	rec := newCloseNotifyingRecorder()
	rw := NewResponseWriter(rec)
	closed := false
	notifier := rw.(http.CloseNotifier).CloseNotify()
	rec.close()
	select {
	case <-notifier:
		closed = true
	case <-time.After(time.Second):
	}
	assert.Equals(t, true, closed)
}

func TestResponseWriterFlusher(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	_, ok := rw.(http.Flusher)
	assert.Equals(t, true, ok)
}
