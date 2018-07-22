package session

import (
	"fmt"
	"net/http"
	"time"

	"github.com/c4milo/handlers/internal"
	"github.com/pkg/errors"
)

type handler struct {
	name   string
	domain string
	maxAge int
	keys   []string
	store  Store
}

// http://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
type option func(*handler)

// WithName allows setting the cookie name for storing the session.
func WithName(n string) option {
	return func(h *handler) {
		h.name = n
	}
}

// WithDomain allows setting the cookie name for storing the session.
func WithDomain(d string) option {
	return func(h *handler) {
		h.domain = d
	}
}

// WithStore sets a specific backing store for session data. By default, the built-in Cookie Store is used.
func WithStore(store Store) option {
	return func(h *handler) {
		h.store = store
	}
}

// WithSecretKey allows to configure the secret key to encrypt and authenticate the session data.
// Key rotation is supported, the left-most key is always the current key.
func WithSecretKey(k ...string) option {
	return func(h *handler) {
		h.keys = k
	}
}

// WithMaxAge allows to set the duration of the session.
func WithMaxAge(d time.Duration) option {
	return func(h *handler) {
		h.maxAge = int(d.Seconds())
		// h.expires = time.Now().Add(d)
	}
}

// Load loads the session, either form the built-in cookie store or an external Store
func (h *handler) Load(r *http.Request) (*Session, error) {
	s := new(Session)
	cookie, err := r.Cookie(h.name)
	if err != nil {
		// No session cookie is found, we create and initialize a new one.
		s.Cookie = &http.Cookie{
			Name:     h.name,
			MaxAge:   h.maxAge,
			Domain:   h.domain,
			HttpOnly: true,
		}
		s.data = make(map[interface{}]interface{})
		s.isDirty = true

		// When external stores are configured, we want to store the session ID in
		// the cookie to be able to store and retrieve its data.
		if h.store != nil {
			s.Value = genID(idSize)
		}
	} else {
		s.Cookie = cookie
	}

	s.Keys = h.keys

	if r.TLS != nil {
		s.Secure = true
	}

	data := []byte(s.Value)
	if h.store != nil {
		sessionID := s.Value
		data, err = h.store.Load(sessionID)
		if err != nil {
			return s, errors.Wrapf(err, "failed loading session ID: %s", sessionID)
		}
	}

	if err := s.decode(data); err != nil {
		return s, err
	}

	return s, nil
}

// Save persist session data either on the built-in cookie store or an external Store.
// When external store is used, the cookie's value contains the session ID.
func (h *handler) Save(w http.ResponseWriter, s *Session) error {
	if !s.isDirty {
		return nil
	}

	defer http.SetCookie(w, s.Cookie)

	// If session was destroyed by user, make sure the destroy operation,
	// from external Store, is also invoked.
	if s.MaxAge == -1 && h.store != nil {
		return h.store.Destroy(s.Value)
	}

	data, err := s.encode()
	if err != nil {
		return err
	}

	if h.store != nil {
		return h.store.Save(s.Value, data)
	}

	s.Value = string(data[:])
	return nil
}

// Handler verifies and creates new sessions. If a session is found and valid,
// it is attached to the Request's context for further modification or retrieval by other
// handlers. Sessions are automatically saved before sending the response.
func Handler(h http.Handler, opts ...option) http.Handler {
	sh := new(handler)
	sh.name = "hs"
	sh.maxAge = 86400 // 1 day

	for _, opt := range opts {
		opt(sh)
	}

	if len(sh.keys) == 0 {
		panic("session: at least one secret key is required")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := sh.Load(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal error: %#v", err), http.StatusInternalServerError)
		}
		ctx := newContext(r.Context(), session)
		res := internal.NewResponseWriter(w)
		res.Before(func(w internal.ResponseWriter) {
			sh.Save(w, session)
		})

		h.ServeHTTP(res, r.WithContext(ctx))
	})
}
