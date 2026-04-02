# Wails Desktop App

This is the supported desktop direction for Windows and Linux.

## Layout

- `desktop/wails/main.go`: Wails entrypoint and desktop window configuration.
- `desktop/wails/app.go`: Wails backend adapter over `internal/appcore`, including the runtime update pump.
- `desktop/wails/frontend/src/`: TypeScript desktop UI source.
- `desktop/wails/frontend/src/ui/`: explicit setup, lobby, and game screen modules plus shared panel surfaces.
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
npm run test:frontend --prefix desktop/wails
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
- Frontend regression suite only: `make wails-frontend-test`

Artifacts are expected under `bin/gui/wails` and follow `docs/BINARY_NAMING.md`.

## Release Bar

Windows/Linux desktop releases are only considered ready when all of the following are true:

- No silent stalls: every mutating action ends in a reconciled snapshot, a visible error, or a timeout-driven recovery state.
- Offline 2-player and 4-player matches start, play, finish, and reset cleanly.
- Direct and relay online flows cover host, join, start, play, chat, host vote, replacement invite, and leave/reset.
- Disconnect, reconnect, and failover surfaces stay visible in the primary UI without requiring diagnostics.
- Setup, lobby, and game screens remain visually balanced at laptop-width (`1280px`), standard desktop, and larger monitor sizes.
- The Wails frontend rebuilds from `frontend/src` and the Go adapter tests pass with `go test -tags=wails .`.

## Notes

- `internal/appcore` remains the only runtime contract owner. The Wails layer renders `SnapshotBundle`, dispatches intents, and emits runtime updates to the desktop UI.
- The desktop UI is organized around explicit `setup`, `lobby`, and `game` screens so render failures and recovery states stay scoped instead of breaking the whole app.
- Production readiness for the desktop client means no silent stalls: every mutating action must resolve into a visible error, a reconciled snapshot, or a timeout-driven recovery path.
- The Go entrypoints are build-tagged so the main module continues to build and test without Wails unless `-tags=wails` is requested.
- The Wails CLI is not vendored in this repository. The current repo expects a local `wails` installation when running desktop-specific build targets.
