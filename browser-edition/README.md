# Truco - Browser Edition

A runtime-backed browser client for Truco Paulista. The PHP frontend talks to the Go HTTP API, and the API now delegates game and online session behavior to `internal/appcore`, the same shared runtime used by the native shells.

## Prerequisites
- Go 1.22+ (for running the HTTP API).
- A modern web browser.
- (Optional) PHP 8.1+ if you intend to test the legacy PHP web proxy wrappers in `php/`.

## Installation & Build Instructions

For standard local gameplay, run the Go HTTP API and the PHP frontend:

1. **Run the API**
   From the root project directory, run:
   ```bash
   go run browser-edition/cmd/httpapi/main.go
   ```

2. **Serve the PHP frontend**
   From `browser-edition/php`, serve the PHP app with your preferred server, for example:
   ```bash
   php -S 127.0.0.1:9080
   ```

3. **Play the Game**
   Open your browser and navigate to:
   ```
   http://127.0.0.1:9080
   ```

## Development
To modify the browser UI, edit the PHP/CSS files inside `browser-edition/php`.
