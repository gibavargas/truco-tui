# Browser Edition

The browser client is a PHP frontend backed by a local Go HTTP API. The API delegates game and online session behavior to `internal/appcore`, so browser behavior stays aligned with the TUI and native clients.

## Structure

- `cmd/httpapi`: local API server used by the browser frontend
- `php/`: editable development frontend
- `dist/`: packaged output produced by the browser build script
- `scripts/build-web.sh`: packaging script

## Requirements

- Go 1.24+
- PHP 8.1+
- A modern browser

## Local Development

Start the API:

```bash
go run ./browser-edition/cmd/httpapi
```

In another terminal, serve the PHP frontend:

```bash
php -S 127.0.0.1:9080 -t browser-edition/php
```

Open:

```text
http://127.0.0.1:9080/index.php
```

## Packaged Build

Build the distributable browser edition:

```bash
make browser
```

This script:

- compiles the Go API into `browser-edition/dist/truco-api`
- copies the PHP frontend into `browser-edition/dist/`
- validates that `dist/` contains only the copied PHP tree plus the compiled API binary

Serve the packaged output with:

```bash
php -S 127.0.0.1:8080 -t browser-edition/dist
```

Then open:

```text
http://127.0.0.1:8080/index.php
```

## API Behavior

The browser frontend uses session-scoped POST endpoints under `/api/<action>`. The server creates an in-memory runtime per browser session and returns snapshot bundles plus drained events as needed.

Non-OK responses also include `error_code` so the browser adapter can distinguish transport failures from runtime failures without parsing localized text.

Notable actions include:

- `createSession`
- `setLocale`
- `startGame`
- `startOnlineHost`
- `joinOnline`
- `startOnlineMatch`
- `sendChat`
- `sendHostVote`
- `requestReplacementInvite`
- `pullOnlineEvents`

## Notes

- `dist/` is generated output and can be rebuilt at any time from `php/` and `cmd/httpapi`.
- `make verify-browser-dist` validates the generated `dist/` tree without rebuilding it.
- Browser parity expectations are defined in `docs/PARITY.md`.
