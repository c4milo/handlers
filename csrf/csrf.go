// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.



package csrf

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/xsrftoken"
)

// SessionManager is used to retrieve a session ID as it is used for hashing the CSRF token.
type SessionManager interface {
	ID() string
}

// handler is a private struct which contains the handler's configurable options.
type handler struct {
	name    string
	domain  string
	secret  string
	session SessionManager
}

// Name allows configuring the CSRF cookie name.
func Name(n string) option {
	return func(h *handler) {
		h.name = n
	}
}

// Secret configures the secret cryptographic key for signing the token.
func Secret(s string) option {
	return func(h *handler) {
		h.secret = s
	}
}

// Session configures the session handler that is going to be used to retrieve the "userID" key value.
func Session(s SessionManager) option {
	return func(h *handler) {
		h.session = s
	}
}

// Domain configures the domain under which the CSRF cookie is going to be set.
func Domain(d string) option {
	return func(h *handler) {
		h.domain = d
	}
}

var (
	errInvalidCSRFToken = errors.New("Invalid CSRF token.")
)

// http://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
type option func(*handler)

// Handler checks Origin header first, if not set or has value "null" it validates using
// a HMAC CSRF token. For enabling Single Page Applications to send the XSRF cookie using
// async HTTP requests, use CORS and make sure Access-Control-Allow-Credential is enabled.
func Handler(h http.Handler, opts ...option) http.Handler {
	// Sets default options
	csrf := &handler{
		name: "xt",
	}

	for _, opt := range opts {
		opt(csrf)
	}

	if csrf.secret == "" {
		panic("csrf: A secret key must be provided")
	}

	if csrf.session == nil {
		panic("csrf: A session ID provider is required")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Details about Origin header can be found at https://wiki.mozilla.org/Security/Origin
		originValue := r.Header.Get("Origin")
		if originValue != "" && originValue != "null" {
			originURL, err := url.Parse(originValue)
			if err == nil && originURL.Host == r.Host {
				h.ServeHTTP(w, r)
				return
			}
		}

		sessionID := csrf.session.ID()
		if sessionID == "" {
			log.Println("csrf: Skipped setting token as there is not a current session.")
			h.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(csrf.name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if !xsrftoken.Valid(cookie.Value, csrf.secret, sessionID, "Global") {
			http.Error(w, errInvalidCSRFToken.Error(), http.StatusForbidden)
			return
		}

		token := xsrftoken.Generate(csrf.secret, sessionID, "Global")
		cookie = &http.Cookie{
			Name:     csrf.name,
			Value:    token,
			Path:     "/",
			Domain:   csrf.domain,
			Expires:  time.Now().Add(xsrftoken.Timeout),
			MaxAge:   int(xsrftoken.Timeout.Seconds()),
			Secure:   true,
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)

		h.ServeHTTP(w, r)
	})
}
