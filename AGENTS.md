# AGENTS.md

## Repo Scope
- Primary app: Go terminal client in `cmd/truco`.
- Relay service: `cmd/truco-relay`.
- Shared native runtime: `cmd/truco-core-ffi`.
- Browser edition: Go HTTP API in `browser-edition/cmd/httpapi` plus the TypeScript static client in `browser-edition/web`.

## Preferred Commands
- Run TUI: `make run` or `go run ./cmd/truco`
- Build host-platform TUI: `make build`
- Run relay locally: `make relay`
- Run tests: `make test` or `go test ./...`
- Clean built artifacts: `make clean`

## Browser Edition
- Build distributable browser bundle: `make browser`
- Browser build script writes `browser-edition/dist`, compiles the Go API to `browser-edition/dist/truco-api`, and bundles the static browser assets there.
- After `make browser`, run locally with:
  - `browser-edition/dist/truco-api`
  - open `http://127.0.0.1:9090/`

## FFI / Native Clients
- Build macOS FFI library: `make ffi` or `make ffi-macos`
- Build Linux FFI library: `make ffi-linux`
- Build Linux GTK client end-to-end: `make linux-gtk`
- Linux GTK app expects `libtruco_core.so` and looks in `TRUCO_CORE_LIB`, `bin/libtruco_core.so`, `native/linux-gtk/lib/libtruco_core.so`, `lib/libtruco_core.so`, and next to the packaged executable
- Flatpak manifest: `native/linux-gtk/dev.truco.Native.yaml`
- Build Linux Flatpak locally: `make flatpak-linux`
- Build Windows FFI DLL: `make ffi-windows`
- Windows portable packaging uses `build-portable.bat`
  - x64 output: `bin/gui/winui/truco-gui-winui-windows-amd64-portable`
  - ARM64 fallback output: `bin/tui/truco-tui-core-windows-arm64-portable.exe`

## Artifact Conventions
- Follow `docs/BINARY_NAMING.md`: `truco-<type>-<client>-<platform>-<arch>[-<variant>]`
- TUI artifacts belong under `bin/tui`
- GUI artifacts belong under `bin/gui/<client>`
- Before rebuilding, check whether the expected artifact already exists and is fresh.

## CI Notes
- CI runs `go mod verify`, `go vet ./...`, `go test ./...`, and `staticcheck ./...`
- Linux native work should also run `cargo test --manifest-path native/linux-gtk/Cargo.toml` when Rust dependencies are available
- If `staticcheck` is needed locally and missing, install it with `go install honnef.co/go/tools/cmd/staticcheck@latest`
