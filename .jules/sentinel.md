## 2024-05-18 - Missing secure session cookie attributes in PHP
**Vulnerability:** The PHP browser edition session cookie lacked `secure`, `httponly`, and `samesite` attributes, potentially exposing session identifiers to XSS and CSRF attacks.
**Learning:** In PHP, `session_set_cookie_params()` must be called before `session_start()` to effectively configure cookie security properties. Hardcoded configurations without checking the request protocol might break local development.
**Prevention:** Always initialize sessions with strict security parameters and conditionally handle the `secure` flag based on `$_SERVER['HTTPS']`.
