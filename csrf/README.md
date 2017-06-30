## CSRF handler
Offers stateless protection against CSRF attacks for Go web applications.

* Checks [Origin header](https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)_Prevention_Cheat_Sheet#Verifying_Same_Origin_with_Standard_Headers) was sent and matches the Host header.
* Falls back to a URL-safe and secure HMAC token stored in a HTTP-only
and secured cookie.
* Protects all HTTP requests that would potentially mutate data: POST, PUT, DELETE and PATCH.
* If you use [CORS](http://www.html5rocks.com/en/tutorials/cors/),
make sure to enable `Access-Control-Allow-Credentials`, so that the cookie containing the HMAC token is sent to
your backend service and can be verified by this handler.
* Allows content to be cacheable by CDNs as the token is sent in a cookie and not on the HTML document.

### Assumptions
* HTTP Origin header is the best way to deflect CSRF attacks, though, some old browsers may not support
it, therefore we provide a fallback to stateless HMAC tokens.
* TLS everywhere has been made possible by https://letsencrypt.org, so this handler only sends the CSRF cookie over TLS.
* Synchronizer Token Pattern is another way of protection, however, this handler offers a simpler and equally effective protection.
* This handler depends on a session or user ID, so you must implement the [Session interface](https://github.com/c4milo/handlers/blob/master/csrf/csrf.go#L15-L17) to allow the handler to retrieve the
session ID from wherever it is being stored.

### Further hardening
To make things a bit more difficult to malicious folks, take a look at defining
your own [Content Security Policy](http://www.html5rocks.com/en/tutorials/security/content-security-policy/)

### References
1. http://www.cs.utexas.edu/~shmat/courses/cs378_spring09/zeller.pdf
2. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
3. http://security.stackexchange.com/questions/91165/why-is-the-synchronizer-token-pattern-preferred-over-the-origin-header-check-to
4. https://bugzilla.mozilla.org/show_bug.cgi?id=446344
5. http://lists.webappsec.org/pipermail/websecurity_lists.webappsec.org/2011-February/007533.html
6. http://stackoverflow.com/questions/24680302/csrf-protection-with-cors-origin-header-vs-csrf-token
7. https://www.owasp.org/index.php/Testing_for_CSRF_(OTG-SESS-005)
8. https://www.fastly.com/blog/caching-uncacheable-csrf-security
9. http://stackoverflow.com/questions/2870371/why-is-jquerys-ajax-method-not-sending-my-session-cookie
