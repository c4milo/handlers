# Golang Session Handler

## Features

* Secured by default through HTTP and secured only cookies, that are also encrypted and authenticated using XSalsa20 and Poly1305.

* Stateless. Since cookies are used as store, session data can't go over 4KB.
