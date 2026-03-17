# Binary Naming Policy

The repository separates terminal and GUI outputs so automation and contributors can quickly identify what has already been built.

## Naming Format

Artifacts follow this pattern:

```text
truco-<type>-<client>-<platform>-<arch>[-<variant>]
```

Where:

- `<type>`: `tui` or `gui`
- `<client>`: consumer identifier such as `core` or `winui`
- `<platform>`: target OS such as `linux`, `darwin`, or `windows`
- `<arch>`: target architecture such as `amd64` or `arm64`
- `<variant>`: optional qualifier such as `portable`

## Output Directories

- TUI artifacts: `bin/tui`
- GUI artifacts: `bin/gui/<client>`

Shared native libraries used by desktop clients are currently built into `bin/` rather than the `bin/tui` or `bin/gui` subtrees:

- macOS: `bin/libtruco_core.dylib`
- Linux: `bin/libtruco_core.so`
- Windows: `bin/truco-core-ffi.dll`

## Current Examples

| Artifact | Path | Notes |
| --- | --- | --- |
| Host TUI | `bin/tui/truco-tui-core-$(go env GOOS)-$(go env GOARCH)` | Host platform build via `make build` |
| Windows TUI x64 | `bin/tui/truco-tui-core-windows-amd64.exe` | Built via `make windows` |
| Windows TUI ARM64 portable | `bin/tui/truco-tui-core-windows-arm64-portable.exe` | Produced by the Windows packaging flow |
| WPF portable bundle | `bin/gui/wpf/truco-gui-wpf-windows-amd64-portable` | Legacy Windows GUI portable output |
| WinUI portable bundle | `bin/gui/winui/truco-gui-winui-windows-amd64-portable` | Portable Windows GUI output |

## Guidance

1. Check whether the expected output already exists before rebuilding.
2. Do not reuse a `<type>-<client>-<platform>-<arch>` tuple for a different artifact.
3. Keep new GUI outputs under `bin/gui/<client>`.
4. Record new variants here when adding them to scripts or CI.
5. Use `make clean` only when a full rebuild is actually needed.
