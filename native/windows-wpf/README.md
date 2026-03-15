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

## Project Layout

- `Models.cs`: snapshot and view models
- `Services/TrucoCoreService.cs`: native bridge and DLL extraction
- `Services/StringProvider.cs`: localization resources
- `ViewModels/MainViewModel.cs`: application state and commands
- `MainWindow.xaml`: main layout
