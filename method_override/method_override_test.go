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
