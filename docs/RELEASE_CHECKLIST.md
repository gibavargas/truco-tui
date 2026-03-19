# Release Checklist

Use this checklist before cutting a release or sharing build artifacts outside local development.

## Validation

- Run `go mod verify`
- Run `go vet ./...`
- Run `go test ./...`
- Run `staticcheck ./...`
- Run `make browser`
- Run `make verify-artifacts`
- For Linux native work, run `cargo test --manifest-path native/linux-gtk/Cargo.toml`

## Artifacts

- Host TUI: `make build`
- Host relay binary: `make build-relay`
- Browser bundle: `make browser`
- macOS FFI: `make ffi-macos`
- Linux FFI and GTK app: `make linux-gtk`
- Windows FFI: `make ffi-windows`
- Windows WinUI portable bundle: `native\windows-winui\build-portable.bat`
- Windows WPF portable bundle, compatibility only: `native\windows-wpf\build-portable.bat`

## Output Review

- Confirm TUI and GUI artifact names follow `docs/BINARY_NAMING.md`
- Confirm `browser-edition/dist` contains only the PHP source tree plus the compiled `truco-api` binary
- Confirm generated build output is not being treated as source-of-truth
- Confirm browser and native adapters still consume the runtime contract from `docs/PARITY.md`

## Smoke Tests

- Offline 2-player match starts and reaches live gameplay
- Offline 4-player match starts and reaches live gameplay
- Host/join direct online session works
- Relay-backed session creation and join work
- Chat works in lobby or match
- Host transfer vote works
- Replacement invite flow enforces runtime preconditions
- Leaving or resetting a session returns to `idle`

## Version Touchpoints

- Verify runtime contract compatibility via `TrucoCoreVersionsJSON`
- If `SnapshotBundle`/`truco.Snapshot` changed, bump `snapshot_schema_version`, rebuild every FFI artifact, and confirm native clients reject stale runtimes
- Verify relay protocol version stays aligned with invite/session code
- Record any public runtime contract changes in `docs/PARITY.md`
