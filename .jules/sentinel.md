## 2026-04-06 - [Insecure Session Management]
**Vulnerability:** The PHP application called `session_start()` without properly configuring secure session cookie attributes.
**Learning:** By default, PHP sessions may not have the `secure`, `httponly`, or `samesite` attributes set. This makes session cookies vulnerable to interception over unencrypted HTTP, access by client-side scripts (XSS), and cross-site request forgery (CSRF).
**Prevention:** Always call `session_set_cookie_params()` before `session_start()`. Set `httponly` to true, `samesite` to 'Lax' or 'Strict', and conditionally set `secure` to true if the request is over HTTPS.
