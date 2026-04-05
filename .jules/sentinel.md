## 2024-04-05 - [PHP Session Security Configuration]
**Vulnerability:** Missing secure session cookie parameters configuration before `session_start()` in PHP. By default, `PHPSESSID` cookie lacks security attributes such as `HttpOnly`, `Secure`, and `SameSite`.
**Learning:** Default PHP session behavior does not inherently protect cookies against XSS (due to lack of HttpOnly) or CSRF (due to lack of SameSite policy). Session fixation or hijacking might be possible over non-HTTPS connections.
**Prevention:** Always explicitly call `session_set_cookie_params()` before `session_start()` to set the `httponly`, `secure`, and `samesite` attributes.
