# Binary Naming Policy

The repository keeps TUI and GUI artifacts in distinct folders so that automated agents and models can recognize which client is already built and avoid redundant work.

## Format
Every produced binary or bundle follows the strict pattern:

```
truco-<type>-<client>-<platform>-<arch>[-<variant>]
```

- `<type>` is either `tui` or `gui`.
- `<client>` names the consumer of the shared runtime (`core`, `winui`, etc.).
- `<platform>` is the target operating system (`linux`, `darwin`, `windows`).
- `<arch>` is the CPU architecture (`amd64`, `arm64`).
- `<variant>` is optional and describes qualifiers such as `portable` or `singlefile`.

TUI artifacts live under `bin/tui`, while GUI bundles are grouped by client under `bin/gui/<client>`.

## Examples

| Client         | Path                                                       | Notes
| -------------- | ---------------------------------------------------------- | ---------------------------------------------
| Host TUI (Linux/macOS) | `bin/tui/truco-tui-core-$(go env GOOS)-$(go env GOARCH)` | Captures host platform/arch via `go env`
| Windows TUI (x64) | `bin/tui/truco-tui-core-windows-amd64.exe`             | Built via `make windows`
| Windows TUI (ARM64 portable) | `bin/tui/truco-tui-core-windows-arm64-portable.exe` | Produced when the WinUI toolchain is unavailable on ARM64
| WinUI GUI (x64 portable) | `bin/gui/winui/truco-gui-winui-windows-amd64-portable` | Output of `build-portable.bat`

## Guidance for models and automation

1. **Check before watching**: Before running `go build` or `dotnet publish`, verify whether the expected path already exists. If the binary at `bin/tui/<name>` or the bundle under `bin/gui/<client>/<name>` is present and fresh, skip recompilation.
2. **Use names to disambiguate**: Never reuse the same `<type>-<client>-<platform>-<arch>` tuple for different projects. When adding new clients, assign a new `<client>` token so existing binaries keep their unique identities.
3. **Clean intentionally**: Use `make clean` to remove `bin/tui` and `bin/gui` when you need a rebuild from scratch.
4. **Document new variants**: If you introduce a variant (for example `singlefile`), append it to the filename and record it here so future tasks know which filename to look for.

Following this policy keeps the repository tidy and prevents agents from re-triggering builds whose outputs already exist.
