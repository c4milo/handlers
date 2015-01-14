package logger

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newRequest(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	return req
}

func TestHandler(t *testing.T) {
	requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	})

	logging := new(bytes.Buffer)
	logHandler := Handler(requestHandler, AppName("test"), Output(logging))

	ts := httptest.NewServer(logHandler)
	defer ts.Close()

	_, err := http.Get(ts.URL)
	expect(t, err, nil)
	refute(t, logging.String(), "")
}
