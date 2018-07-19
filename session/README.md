# Golang Session Handler

Session management for Go web apps

* Uses Go's standard library as much as possible.

* Automatically saves session data at the end of each request-response lifecycle.

* Only makes calls to backing store if session data actually changed.

* Secured by default through HTTP and secured only cookies, that are also encrypted and authenticated using XSalsa20 and Poly1305.

* Cookie Store built-in and use as default.

* Extensible through the implementation of new Stores.
