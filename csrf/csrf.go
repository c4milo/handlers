// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package csrf offers stateless protection against CSRF attacks using
// the HTTP Origin header and falling back to HMAC tokens stored on secured
// and HTTP-only cookies.
package csrf

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/xsrftoken"
)

// handler is a private struct which contains the handler's configurable options.
type handler struct {
	name   string
	domain string
	secret string
	userID string
}

// WithName allows configuring the CSRF cookie name.
func WithName(n string) Option {
	return func(h *handler) {
		h.name = n
	}
}

// WithSecret configures the secret cryptographic key for signing the token.
func WithSecret(s string) Option {
	return func(h *handler) {
		h.secret = s
	}
}

// WithUserID allows to configure a random and unique user ID identifier used to generate the CSRF token.
func WithUserID(s string) Option {
	return func(h *handler) {
		h.userID = s
	}
}

// WithDomain configures the domain under which the CSRF cookie is going to be set.
func WithDomain(d string) Option {
	return func(h *handler) {
		h.domain = d
	}
}

var (
	// We are purposely being ambiguous on the HTTP error messages to avoid giving clues to potential attackers
	// other than 403 Forbidden messages
	errForbidden = "Forbidden"
	// Development time messages
	errSecretRequired = errors.New("csrf: a secret key must be provided")
)

// Option implements http://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
type Option func(*handler)

// Handler checks Origin header first, if not set or has value "null" it validates using
// a HMAC CSRF token. For enabling Single Page Applications to send the XSRF cookie using
// async HTTP requests, use CORS and make sure Access-Control-Allow-Credential is enabled.
func Handler(h http.Handler, opts ...Option) http.Handler {
	// Sets default options
	csrf := &handler{
		name: "xt",
	}

	for _, opt := range opts {
		opt(csrf)
	}

	if csrf.secret == "" {
		panic(errSecretRequired)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Re-enables browser's XSS filter if it was disabled
		w.Header().Set("x-xss-protection", "1; mode=block")

		// Set the token on the response to GET and HEAD requests
		switch r.Method {
		case "GET":
		case "HEAD":
		case "OPTIONS":
			setToken(w, csrf.name, csrf.secret, csrf.userID, csrf.domain)
			h.ServeHTTP(w, r)
			return
		}

		// Verify using origin header first
		// Details about Origin header can be found at https://wiki.mozilla.org/Security/Origin
		originValue := r.Header.Get("origin")
		if originValue != "" {
			originURL, err := url.ParseRequestURI(originValue)
			if err == nil && originURL.Host == r.Host {
				setToken(w, csrf.name, csrf.secret, csrf.userID, csrf.domain)
				h.ServeHTTP(w, r)
				return
			}
			// log.Printf("csrf: %+v\n", err)
		}

		// If origin is not supported or came back empty or null, verify cookie instead.
		cookie, err := r.Cookie(csrf.name)
		if err != nil {
			http.Error(w, errForbidden, http.StatusForbidden)
			return
		}

		if !xsrftoken.Valid(cookie.Value, csrf.secret, csrf.userID, "Global") {
			http.Error(w, errForbidden, http.StatusForbidden)
			return
		}

		setToken(w, csrf.name, csrf.secret, csrf.userID, csrf.domain)
		h.ServeHTTP(w, r)
	})
}

func setToken(w http.ResponseWriter, name, secret, userID, domain string) {
	token := xsrftoken.Generate(secret, userID, "Global")

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		Domain:   domain,
		Expires:  time.Now().Add(xsrftoken.Timeout),
		MaxAge:   int(xsrftoken.Timeout.Seconds()),
		Secure:   true,
		HttpOnly: true,
	})
}
