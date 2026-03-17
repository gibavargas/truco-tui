# Windows WPF Client

This is a legacy Windows GUI client built with WPF and .NET 8. It also integrates with the shared Go runtime through `truco-core-ffi.dll`.

## Status

The WinUI client in `native/windows-winui` is the primary Windows shell. Keep this WPF client documented as a secondary or compatibility-oriented option unless the project explicitly promotes it again.

## Requirements

- Windows 10 or Windows 11
- .NET 8 SDK
- Go 1.24+

## Build

Build the shared DLL from the repository root:

```bash
make ffi-windows
```

Publish the WPF application:

```bash
dotnet publish native/windows-wpf/TrucoWPF.csproj -c Release -o bin/gui/wpf/truco-gui-wpf-windows-amd64
```

Portable single-file build:

```bash
dotnet publish native/windows-wpf/TrucoWPF.csproj -c Release -r win-x64 --self-contained true -p:PublishSingleFile=true -o bin/gui/wpf/truco-gui-wpf-windows-amd64-portable
```

Client-specific helper:

```bash
native\windows-wpf\build-portable.bat
```

## Project Layout

- `Models.cs`: snapshot and view models
- `Services/TrucoCoreService.cs`: native bridge and portable DLL probing
- `Services/StringProvider.cs`: localization resources
- `ViewModels/MainViewModel.cs`: application state and commands
- `MainWindow.xaml`: main layout
- `build-portable.bat`: WPF portable packaging flow

## Portable Runtime Layout

The WPF client now loads the shared runtime from app-local paths instead of extracting it to a temp folder.
At startup it checks, in order:

- `TRUCO_CORE_LIB`
- `<app>\truco-core-ffi.dll`
- `<app>\lib\truco-core-ffi.dll`
- `<cwd>\bin\truco-core-ffi.dll`
- `<cwd>\native\windows-wpf\lib\truco-core-ffi.dll`

Portable publishes place the Go runtime at `lib\truco-core-ffi.dll` inside the bundle so the application can be copied as a self-contained directory.
