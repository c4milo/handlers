package session

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/hooklift/assert"
)

func TestHandler(t *testing.T) {
	requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := FromContext(r.Context())
		session.Set("blah", "camilo")
		value := session.Get("blah")
		fmt.Fprintf(w, "Hello %s!", value)
	})

	sessionHandler := Handler(requestHandler, WithSecretKey(
		"secret", "old1", "old2", "old3",
	))

	ts := httptest.NewServer(sessionHandler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Ok(t, err)

	dump, err := httputil.DumpResponse(resp, true)
	assert.Ok(t, err)

	fmt.Printf("%q", dump)
}
