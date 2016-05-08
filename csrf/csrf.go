package csrf

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/xsrftoken"
)

// Session is used to retrieve userID from the session as it is used as part of the hashing
// of the CSRF token.
type Session interface {
	ID() string
}

// handler is a private struct which contains the handler's configurable options.
type handler struct {
	name    string
	domain  string
	secret  string
	session Session
}

// SetName allows configuring the CSRF cookie name.
func SetName(n string) option {
	return func(h *handler) {
		h.name = n
	}
}

// SetSecret configures the secret cryptographic key for signing the token.
func SetSecret(s string) option {
	return func(h *handler) {
		h.secret = s
	}
}

// SetSession configures the session handler that is going to be used to retrieve the "userID" key value.
func SetSession(s Session) option {
	return func(h *handler) {
		h.session = s
	}
}

// SetDomain configures the domain under which the CSRF cookie is going to be set.
func SetDomain(d string) option {
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

		if csrf.session == nil {
			panic("csrf: A session ID provider is required")
		}

		sessionID := csrf.session.ID()
		if sessionID == "" {
			log.Println("csrf: Skipped setting token as there is no a current session.")
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
