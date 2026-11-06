# Project Status

This file tracks the latest documentation-aligned project status. Last updated: 2026-11-05.

## Clients

| Client | Location | Status |
| --- | --- | --- |
| Terminal (Bubble Tea) | `cmd/truco`, `internal/ui` | Supported |
| Browser edition | `browser-edition/` (Go HTTP API + TypeScript SPA) | Supported; live at truco.jvguidi.com.br. The PHP implementation is gone, the packaged build is a single Go binary plus static assets. |
| Wails desktop | `desktop/wails` | Supported desktop direction for Windows and Linux |
| macOS native | `native/macos` | Supported |
| Linux GTK / Windows WinUI | `native/linux-gtk`, `native/windows-winui` | Deprecated in favor of Wails |
| Ebitengine 2D client | - | Removed from the repo (along with its `internal/clientparity` source-scan test) |
| Windows WPF | - | Removed from the repo; docs no longer reference it |

## Runtime and Services

- Shared runtime in `internal/appcore` (rules, sessions, snapshots, events). Client behavior contract lives in `docs/PARITY.md`.
- CPU opponents use a C++17 strategy engine (`native/ai-engine`, `libtruco_ai`) linked through CGO, including a `HybridStrategy` that blends personalities per CPU seat. CGO-enabled Go builds/tests need `make -C native/ai-engine` first; `CGO_ENABLED=0` builds (the CI cross-compile matrix) use the pure-Go fallback in `internal/truco/cpu_nocgo.go`, which has no C++ personalities.
- Relay (`cmd/truco-relay`): HTTPS control plane, QUIC data plane, and a TLS-over-TCP fallback (`TRUCO_RELAY_TCP_ADDR`, default `:9445`) for UDP-blocked networks. Validated with `scripts/e2e-tcp-smoke`.
- Browser visual identity: "TrucoON" redesign (flat, orange/green, Montserrat).

## CI

- `.github/workflows/ci.yml` jobs: `test` (build C++ engine, `go mod verify`, vet, test, staticcheck, Wails CLI, `make browser`, `make verify-artifacts`), `linux-native` (GTK Rust tests, Flatpak manifest check), `build` (cross-platform TUI and relay binaries).
- `test` runs on Ubuntu 22.04 because Wails v2 needs `webkit2gtk-4.0`. Node 20 is provisioned for the browser and Wails frontends, and staticcheck is pinned.
- The `build` job cross-compiles with `CGO_ENABLED=0`; it relies on the `!cgo` fallback because `cpu_cpp.go` needs CGO. Those release binaries therefore use the Go CPU heuristic, not the C++ engine.

## Open Items

- Confirm a green CI run on `main` after the latest workflow changes (it was red from 2026-08-19 until the AI-engine build step was added).
- Rotate any credentials that were in the removed `config.yaml`; it is still present in git history.
- Keep `docs/PARITY.md` in sync when the runtime contract changes (`snapshot_schema_version`).
