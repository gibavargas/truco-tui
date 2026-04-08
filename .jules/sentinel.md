
## 2024-05-30 - Insecure Session Management Fix
**Vulnerability:** Missing secure attributes on PHP session cookies in the browser-edition application.
**Learning:** In PHP, `session_set_cookie_params()` must be called before `session_start()` to enforce secure attributes like `secure`, `httponly`, and `samesite` on session cookies. To support both local dev environments (HTTP) and production (HTTPS), the `secure` flag needs to be set dynamically based on `isset($_SERVER['HTTPS']) && $_SERVER['HTTPS'] === 'on'`.
**Prevention:** Implement the dynamic session cookie parameter pattern at the bootstrap layer of the application before initializing the session.
