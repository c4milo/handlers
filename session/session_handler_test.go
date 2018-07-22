package session

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hooklift/assert"
)

func TestHandlerSave(t *testing.T) {
	requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := FromContext(r.Context())
		session.Set("blah", "gophersito")
		value := session.Get("blah")
		fmt.Fprintf(w, "Hello %s!", value)
	})

	sessionHandler := Handler(requestHandler, WithSecretKey(
		"new",
	))

	ts := httptest.NewServer(sessionHandler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Ok(t, err)
	assert.Cond(t, len(resp.Cookies()) > 0, "no session cookie found")

	cookie := resp.Cookies()[0]
	s := new(Session)
	s.Cookie = cookie
	s.Keys = []string{"new"}
	s.decode([]byte(cookie.Value))
	assert.Equals(t, "gophersito", s.Get("blah"))
}
