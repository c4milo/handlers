// Package session offers a HTTP handler to manage web sessions using
// encrypted and authenticated cookies as well as pluggable backing stores.
package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	// https://www.owasp.org/index.php/Insufficient_Session-ID_Length
	idSize = 32 // 256 bits
)

// Store defines the contract for implementing different session data stores.
type Store interface {
	// Load retrieves opaque session data from backing store
	Load(id string) ([]byte, error)
	// Saves persists opaque session data to backing store
	Save(id string, data []byte) error
	// Destroy removes the session altogether from backing store
	Destroy(id string) error
}

// Session represents a secure session cookie. By default, it stores session data in the
// session cookie. If a Store is provided, only the session ID is stored in the session cookie.
type Session struct {
	*http.Cookie
	// keys is the key used to encrypt and authenticate the session cookie's value
	Keys []string
	// data is where the session data is temporarly loaded to for manipulation,
	// during the request-response lifecycle.
	data map[interface{}]interface{}
	// isDirty determines whether the session data must be save or not.
	isDirty bool
}

// New returns a new Session
func New() *Session {
	session := Session{
		data:   make(map[interface{}]interface{}),
		Cookie: new(http.Cookie),
	}
	return &session
}

// Set assigns a value to a specific key.
func (s *Session) Set(key string, value interface{}) error {
	s.isDirty = true

	s.data[key] = value
	return nil
}

// Get retrieves the given key's value from the session store.
func (s *Session) Get(key string) interface{} {
	return s.data[key]
}

// Delete removes the given key's value from the session store.
func (s *Session) Delete(key string) error {
	s.isDirty = true

	delete(s.data, key)
	return nil
}

// Destroy signals the user's browser to remove the session cookie.
func (s *Session) Destroy() {
	s.MaxAge = -1
	s.Expires = time.Now()
	s.data = make(map[interface{}]interface{})
	s.isDirty = true
}

// Encode encrypts and serializes the session cookie's data.
func (s *Session) Encode() ([]byte, error) {
	if len(s.data) == 0 {
		return nil, nil
	}

	if len(s.Keys) == 0 {
		return nil, errors.New("at least one encryption key is required")
	}

	msg, err := msgpack.Marshal(s.data)
	if err != nil {
		return nil, errors.Wrapf(err, "failed encoding session data")
	}

	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, errors.Wrapf(err, "failed generating nonce value")
	}

	var key [32]byte
	copy(key[:], s.Keys[0])

	box := secretbox.Seal(nonce[:], msg, &nonce, &key)

	data := make([]byte, base64.RawStdEncoding.EncodedLen(len(box)))
	base64.RawStdEncoding.Encode(data, box)

	return data, nil
}

// Decode decrypts, authenticates and deserializes cookie's session data.
func (s *Session) Decode(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	box := make([]byte, base64.RawStdEncoding.DecodedLen(len(data)))
	if _, err := base64.RawStdEncoding.Decode(box, data); err != nil {
		return errors.Wrapf(err, "failed decoding session data")
	}

	var nonce [24]byte
	var msg []byte
	var key [32]byte
	var ok bool
	copy(nonce[:], box[:24])

	for _, k := range s.Keys {
		copy(key[:], k)
		msg, ok = secretbox.Open(nil, box[24:], &nonce, &key)
		if ok {
			if err := msgpack.Unmarshal(msg, &s.data); err != nil {
				return errors.Wrapf(err, "failed decoding session data")
			}
			return nil
		}
	}

	return errors.New("failed decrypting session data")
}

// sessionKey is the key used to store the session instance in the request's context.
type sessionKey struct{}

// newContext returns a new context with the provided session inside.
func newContext(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionKey{}, s)
}

// FromContext extracts the session from the given context.
func FromContext(ctx context.Context) (s *Session, ok bool) {
	s, ok = ctx.Value(sessionKey{}).(*Session)
	return
}

// genID returns a random string
func genID(size int) string {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", b)
}
