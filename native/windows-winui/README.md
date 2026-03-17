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

This produces `bin/truco-core-ffi.dll`, which the client copies into build/publish output and also embeds as a fallback.

## Publish the Portable App

Preferred packaging entrypoint on Windows:

```bat
native\windows-winui\build-portable.bat
```

Equivalent direct publish command:

```bash
dotnet publish native/windows-winui/TrucoWinUI.csproj -c Release -r win-x64 --self-contained true -p:PublishSingleFile=false -o bin/gui/winui/truco-gui-winui-windows-amd64-portable
```

## ARM64 Note

The current Go toolchain does not support `-buildmode=c-shared` for Windows ARM64 in this setup. The repository’s Windows packaging flow therefore falls back to the TUI binary on ARM64 instead of building the WinUI client natively.

## Runtime Resolution

At startup the WinUI client resolves the shared runtime in this order:

1. `TRUCO_CORE_LIB`
2. `truco-core-ffi.dll` next to the executable
3. `lib\truco-core-ffi.dll` under the executable directory
4. `bin\truco-core-ffi.dll` from the repository root when running from source
5. the embedded fallback copy extracted to the local runtime cache

The client also validates `TrucoCoreVersionsJSON` before creating a runtime handle, so portable bundles fail fast if the DLL and frontend drift out of sync.

## Notes

- The main application state lives in `ViewModels/AppShellViewModel.cs`.
- The project file is configured for self-contained folder publication so the Windows App SDK and the Go runtime stay app-local.
- Binary naming follows `docs/BINARY_NAMING.md`.
