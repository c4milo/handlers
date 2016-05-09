// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package csrf

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hooklift/assert"
	"golang.org/x/net/xsrftoken"
)

var requestHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello world.")
})

func TestSecretRequired(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Equals(t, errSecretRequired, r)
		}
	}()

	handler := Handler(requestHandler)
	ts := httptest.NewServer(handler)
	defer ts.Close()
}

func TestSessionRequired(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Equals(t, errSessionRequired, r)
		}
	}()

	handler := Handler(requestHandler, Secret("my secret!"))
	ts := httptest.NewServer(handler)
	defer ts.Close()
}

type SessionImpl struct{}

func (s *SessionImpl) ID() string {
	return "my ID!"
}

func TestDomainRequired(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Equals(t, errDomainRequired, r)
		}
	}()

	handler := Handler(requestHandler, Session(new(SessionImpl)), Secret("my secret!"))
	ts := httptest.NewServer(handler)
	defer ts.Close()
}

func TestCSRFProtection(t *testing.T) {
	session := new(SessionImpl)
	handler := Handler(requestHandler, Session(session), Secret("my secret!"), Domain("localhost"), Name("_csrf"))
	okTS := httptest.NewServer(handler)
	defer okTS.Close()

	cookie := &http.Cookie{
		Name:     "_csrf",
		Value:    xsrftoken.Generate("my secret!", session.ID(), "Global"),
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(xsrftoken.Timeout),
		MaxAge:   int(xsrftoken.Timeout.Seconds()),
		Secure:   true,
		HttpOnly: true,
	}

	expectedBody := "Hello world."
	tests := []struct {
		origin     string
		body       string
		statusCode int
		cookies    int
		cookie     *http.Cookie
	}{
		{okTS.URL, expectedBody, http.StatusOK, 0, nil},
		{"", expectedBody, http.StatusOK, 1, cookie},
		{"null", expectedBody, http.StatusOK, 1, cookie},
		{"", errForbidden + "\n", http.StatusForbidden, 0, nil},
	}

	for _, tt := range tests {
		//fmt.Printf("# %d\n", tn)
		req, err := http.NewRequest("POST", okTS.URL, nil)
		assert.Ok(t, err)

		req.Header.Set("origin", tt.origin)
		if tt.cookie != nil {
			req.AddCookie(tt.cookie)
		}
		resp, err := http.DefaultClient.Do(req)
		assert.Ok(t, err)

		body, err := ioutil.ReadAll(resp.Body)
		assert.Ok(t, err)
		defer resp.Body.Close()
		assert.Equals(t, tt.body, string(body[:]))
		assert.Equals(t, tt.statusCode, resp.StatusCode)
		assert.Equals(t, tt.cookies, len(resp.Cookies()))

		for _, c := range resp.Cookies() {
			assert.Equals(t, "_csrf", c.Name)
			assert.Cond(t, c != cookie, "csrf cookie has to be different per request")
		}
	}
}
