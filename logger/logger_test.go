// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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
