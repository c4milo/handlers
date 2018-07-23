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

func TestCSRFProtection(t *testing.T) {
	opts := []Option{
		WithUserID("my user ID!"),
		WithSecret("my secret!"),
		WithName("_csrf"),
	}

	handler := Handler(requestHandler, opts...)
	okTS := httptest.NewServer(handler)
	defer okTS.Close()

	cookie := &http.Cookie{
		Name:     "_csrf",
		Value:    xsrftoken.Generate("my secret!", "my user ID!", "Global"),
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(xsrftoken.Timeout),
		MaxAge:   int(xsrftoken.Timeout.Seconds()),
		Secure:   true,
		HttpOnly: true,
	}

	expectedBody := "Hello world."
	tests := []struct {
		desc       string
		origin     string
		body       string
		statusCode int
		cookies    int
		cookie     *http.Cookie
	}{
		{
			"it should accept mutating request from same origin",
			okTS.URL, expectedBody, http.StatusOK, 1, nil,
		},
		{
			"it should accept mutating request if no origin and token is found in cookie",
			"", expectedBody, http.StatusOK, 1, cookie,
		},
		{
			"it should accept mutating request origin is null but a csrf token is found in cookie",
			"null", expectedBody, http.StatusOK, 1, cookie,
		},
		{
			"it should reject request if not origin and no csrf cookie is found",
			"", errForbidden + "\n", http.StatusForbidden, 0, nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			req, err := http.NewRequest("POST", okTS.URL, nil)
			assert.Ok(t, err)

			req.Header.Set("origin", tt.origin)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			resp, err := http.DefaultClient.Do(req)
			assert.Ok(t, err)
			assert.Equals(t, tt.statusCode, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			assert.Ok(t, err)
			defer resp.Body.Close()

			assert.Equals(t, tt.body, string(body[:]))
			//fmt.Printf("%+v\n", resp.Cookies())
			assert.Equals(t, tt.cookies, len(resp.Cookies()))

			for _, c := range resp.Cookies() {
				assert.Equals(t, "_csrf", c.Name)
				assert.Cond(t, c != cookie, "csrf cookie has to be different per request")
			}
		})
	}
}
