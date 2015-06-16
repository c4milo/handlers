package logger

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hooklift/assert"
)

func TestHandler(t *testing.T) {
	requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	})

	logging := new(bytes.Buffer)
	logHandler := Handler(requestHandler, AppName("test"), Output(logging))

	ts := httptest.NewServer(logHandler)
	defer ts.Close()

	_, err := http.Get(ts.URL)
	assert.Ok(t, err)
	assert.Cond(t, logging.String() != "", "Log output should not be empty.")
}
