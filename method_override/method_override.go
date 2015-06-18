// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package methodoverride

import (
	"net/http"
	"strings"
)

// Handler adds support for overriding the HTTP method, especially for clients
// that do not support HTTP methods other than GET and POST.
//
// Examples for clients calling an API using this handler or middleware:
// <form method="POST" action="/resource">
//   <input type='hidden' name='_method' value='PATCH' />
//   <button type="submit">Delete resource</button>
// </form>
//
// curl -n -X POST https://example.com/resource/$ID_OR_NAME \
// -H "Content-Type: application/json" \
// -H "HTTP-Method-Override: PATCH" \
// -d '{
//   "example": "foobar"
// }'
func Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			m := r.FormValue("_method")
			if m == "" {
				m = r.Header.Get("HTTP-Method-Override")
			}

			m = strings.ToUpper(m)

			if m == "PATCH" || m == "PUT" || m == "DELETE" {
				r.Method = m
			}
		}
		h.ServeHTTP(w, r)
	})
}
