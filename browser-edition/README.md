# Browser Edition

The browser client is now a TypeScript single-page app backed by the local Go HTTP API. The API delegates game and online session behavior to `internal/appcore`, so browser behavior stays aligned with the TUI and native clients while the packaged build runs without PHP.

## Structure

- `cmd/httpapi`: local API server used by the browser frontend
- `web/`: editable TypeScript and CSS frontend source
- `public/`: static icons copied into the packaged app
- `dist/`: packaged output produced by the browser build script
- `scripts/build-web.sh`: packaging script

## Requirements

- Go 1.24+
- Node.js 20+
- A modern browser

## Local Development

Install browser dependencies once:

```bash
npm install --prefix browser-edition
```

Build the browser app:

```bash
npm run build --prefix browser-edition
```

Then start the Go server:

```bash
go run ./browser-edition/cmd/httpapi
```

Open:

```text
http://127.0.0.1:9090/
```

## Packaged Build

Build the distributable browser edition:

```bash
make browser
```

This script:

- compiles the Go API into `browser-edition/dist/truco-api`
- bundles the TypeScript frontend into `browser-edition/dist/`
- copies static icons into `browser-edition/dist/`
- validates the expected static layout plus the compiled API binary

Run the packaged output with:

```bash
browser-edition/dist/truco-api
```

Then open:

```text
http://127.0.0.1:9090/
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

- `dist/` is generated output and can be rebuilt at any time from `web/`, `public/`, and `cmd/httpapi`.
- `make verify-browser-dist` validates the generated `dist/` tree without rebuilding it.
- Browser parity expectations are defined in `docs/PARITY.md`.
