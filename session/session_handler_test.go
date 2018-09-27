package session

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hooklift/assert"
)

type testStruct struct {
	Name string
	Test string
}

func init() {
	Register(1, (*testStruct)(nil))
}

func TestSave(t *testing.T) {
	requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := FromContext(r.Context())
		session.Set("blah", "gophersito")
		session.Set("blah2", "deleted")
		session.Set("test", &testStruct{Name: "camilo", Test: "foo"})
		value := session.Get("blah")
		session.Delete("blah2")
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
	s := New()
	s.Cookie = cookie
	s.Keys = []string{"new"}
	err = s.Decode([]byte(cookie.Value))
	assert.Ok(t, err)
	assert.Equals(t, "gophersito", s.Get("blah"))
	assert.Equals(t, nil, s.Get("blah2"))
	assert.Equals(t, &testStruct{Name: "camilo", Test: "foo"}, s.Get("test"))
}

func TestDestroy(t *testing.T) {
	requestHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := FromContext(r.Context())
		session.Set("blah", "gophersito")
		value := session.Get("blah")
		session.Destroy()
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
	assert.Equals(t, "hs", cookie.Name)
	assert.Equals(t, "", cookie.Value)
	assert.Equals(t, -1, cookie.MaxAge)
}
