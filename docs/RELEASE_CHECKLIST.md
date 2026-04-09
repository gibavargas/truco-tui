# Release Checklist

Use this checklist before cutting a release or sharing build artifacts outside local development.

## Validation

- Run `go mod verify`
- Run `go vet ./...`
- Run `go test ./...`
- Run `staticcheck ./...`
- Run `make browser`
- Run `make verify-wails`
- Run `make verify-artifacts`
- For Linux native work, run `cargo test --manifest-path native/linux-gtk/Cargo.toml`

## Artifacts

- Host TUI: `make build`
- Host relay binary: `make build-relay`
- Browser bundle: `make browser`
- Host Wails desktop: `make wails-build`
- Linux Wails desktop: `make wails-build-linux`
- Windows Wails desktop: `make wails-build-windows`
- macOS FFI: `make ffi-macos`
- Linux FFI and GTK app: `make linux-gtk`
- Windows FFI: `make ffi-windows`
- Windows WinUI portable bundle: `native\windows-winui\build-portable.bat`
- Windows WPF portable bundle, compatibility only: `native\windows-wpf\build-portable.bat`

## Output Review

- Confirm TUI and GUI artifact names follow `docs/BINARY_NAMING.md`
- Confirm `browser-edition/dist` contains only the static web bundle plus the compiled `truco-api` binary
- Confirm generated build output is not being treated as source-of-truth
- Confirm browser and native adapters still consume the runtime contract from `docs/PARITY.md`
- Confirm Wails frontend assets rebuild cleanly from `desktop/wails/frontend/src`

## Smoke Tests

- Offline 2-player match starts and reaches live gameplay
- Offline 4-player match starts and reaches live gameplay
- Offline matches can be finished and reset back to `idle` without stale busy state
- Host/join direct online session works through lobby, match start, gameplay, and leave/reset
- Relay-backed session creation and join work through lobby, match start, gameplay, and leave/reset
- Chat works in lobby and match
- Host transfer vote works and updates visible host ownership
- Replacement invite flow works after a real disconnect and replacement join
- Disconnect/reconnect or failover surfaces visible recovery messaging instead of a silent stall
- Leaving or resetting a session returns to `idle`
- Wails desktop window launches with the expected icon, title, and minimum size
- Wails game layout is visually balanced at `1280px` width, a standard desktop window, and a larger monitor
- Wails setup, lobby, and game screens keep network status accessible without diagnostics being mandatory

## Version Touchpoints

- Verify runtime contract compatibility via `TrucoCoreVersionsJSON`
- If `SnapshotBundle`/`truco.Snapshot` changed, bump `snapshot_schema_version`, rebuild every FFI artifact, and confirm native clients reject stale runtimes
- Verify relay protocol version stays aligned with invite/session code
- Record any public runtime contract changes in `docs/PARITY.md`
