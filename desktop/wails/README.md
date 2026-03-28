# Wails Desktop App

This is the supported desktop direction for Windows and Linux.

Current layout:

- `desktop/wails/main.go`: Wails entrypoint, guarded by the `wails` build tag.
- `desktop/wails/app.go`: thin desktop backend over `internal/appcore`.
- `desktop/wails/frontend/`: static frontend shell for the desktop UI.

Notes:

- The Wails CLI is not vendored in this repository.
- The Go entrypoints are build-tagged so the main module continues to build and test without a local Wails installation.
- Once Wails is installed, build this app with the `wails` tag and hook the frontend through the normal Wails workflow.
