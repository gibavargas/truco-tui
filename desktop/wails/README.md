# Wails Desktop App

This is the supported desktop direction for Windows and Linux.

## Layout

- `desktop/wails/main.go`: Wails entrypoint and desktop window configuration.
- `desktop/wails/app.go`: Wails backend adapter over `internal/appcore`, including the runtime update pump.
- `desktop/wails/frontend/src/`: TypeScript desktop UI source.
- `desktop/wails/frontend/dist/`: generated Wails asset bundle.
- `desktop/wails/scripts/build-frontend.mjs`: frontend bundler used by `npm run build` and `wails build`.
- `desktop/wails/build/appicon.png`: source icon used by Wails packaging.

## Local Development

Install frontend dependencies once:

```bash
npm install --prefix desktop/wails
```

Typecheck and build the desktop frontend:

```bash
npm run typecheck --prefix desktop/wails
npm run build --prefix desktop/wails
```

Run the desktop app in Wails dev mode:

```bash
make wails-dev
```

## Build Targets

- Host-platform desktop build: `make wails-build`
- Linux amd64 desktop build: `make wails-build-linux`
- Windows amd64 desktop build: `make wails-build-windows`
- Frontend verification + Wails dry-run release checks: `make verify-wails`

Artifacts are expected under `bin/gui/wails` and follow `docs/BINARY_NAMING.md`.

## Notes

- `internal/appcore` remains the only runtime contract owner. The Wails layer renders `SnapshotBundle`, dispatches intents, and emits runtime updates to the desktop UI.
- The Go entrypoints are build-tagged so the main module continues to build and test without Wails unless `-tags=wails` is requested.
- The Wails CLI is not vendored in this repository. The current repo expects a local `wails` installation when running desktop-specific build targets.
