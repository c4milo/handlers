// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package methodoverride

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hooklift/assert"
)

func TestMethodOverride(t *testing.T) {
	requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equals(t, "PATCH", r.Method)
		fmt.Fprint(w, "Hello")
	})

	methodOverride := Handler(requestHandler)

	ts := httptest.NewServer(methodOverride)
	defer ts.Close()

	req, err := http.NewRequest("POST", ts.URL, nil)
	assert.Ok(t, err)
	req.Header.Set("HTTP-Method-Override", "PATCH")

	resp, err := http.DefaultClient.Do(req)
	assert.Ok(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Ok(t, err)
	assert.Equals(t, "Hello", string(body))
}
