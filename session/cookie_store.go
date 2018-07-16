package session

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"

	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
	"golang.org/x/crypto/nacl/secretbox"
)

// CookieStore implements a secure cookie as session store by encrypting
// and authenticating the cookie's data using XSalsa20 and Poly1305.
type CookieStore struct {
	*http.Cookie
	Keys []string
	// data is used temporarily to store session data, its content is encoded
	// and stored in the cookie's value field when any HTTP handler returns.
	data map[interface{}]interface{}
}

// Open decrypts and decodes session data
func (s *CookieStore) Open() error {
	return s.decode()
}

// Close encrypts and encodes session data
func (s *CookieStore) Close() error {
	return s.encode()
}

// Set stores a specific value in the given key
func (s *CookieStore) Set(key string, value interface{}) error {
	s.data[key] = value
	return nil
}

// Get retrieves the value corresponding to the given key
func (s *CookieStore) Get(key string) interface{} {
	return s.data[key]
}

// Delete removes the value corresponding to the given key
func (s *CookieStore) Delete(key string) error {
	delete(s.data, key)
	return nil
}

// encode encrypts and serializes cookie's data.
func (s *CookieStore) encode() error {
	msg, err := msgpack.Marshal(s.data)
	if err != nil {
		return errors.Wrapf(err, "failed encoding session data")
	}

	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return errors.Wrapf(err, "failed generating nonce value")
	}

	var key [32]byte
	copy(key[:], s.Keys[0])

	box := secretbox.Seal(nonce[:], msg, &nonce, &key)

	s.Value = base64.StdEncoding.EncodeToString(box)

	log.Printf("cookie size: %d bytes", len(s.Name)+len(s.Value))
	return nil
}

// decode decrypts, authenticates and deserializes cookie's data.
func (s *CookieStore) decode() error {
	var box []byte
	if _, err := base64.StdEncoding.Decode(box, []byte(s.Value)); err != nil {
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
