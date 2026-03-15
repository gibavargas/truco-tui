# Windows WinUI Client

This is the primary Windows GUI client. It uses WinUI 3 on .NET 8 and consumes the shared Go runtime through `truco-core-ffi.dll`.

## Requirements

- Windows 11 or Windows 10 with a compatible Windows App SDK environment
- .NET 8 SDK
- Go 1.24+
- Visual Studio 2022 or equivalent Windows build tools

## Build the FFI DLL

From the repository root:

```bash
make ffi-windows
```

This produces `bin/truco-core-ffi.dll`, which the project embeds when present.

## Publish the Portable App

```bash
dotnet publish native/windows-winui/TrucoWinUI.csproj -c Release -r win-x64 --self-contained true -o bin/gui/winui/truco-gui-winui-windows-amd64-portable
```

## ARM64 Note

The current Go toolchain does not support `-buildmode=c-shared` for Windows ARM64 in this setup. The repository’s Windows packaging flow therefore falls back to the TUI binary on ARM64 instead of building the WinUI client natively.

## Notes

- The main application state lives in `ViewModels/AppShellViewModel.cs`.
- The project file is configured for self-contained single-file publication.
- Binary naming follows `docs/BINARY_NAMING.md`.
