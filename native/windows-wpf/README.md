# Truco - WPF Native Windows

This is a WPF (Windows Presentation Foundation) GUI version of the Truco card game.

## Requirements

- Windows 10/11
- .NET 8.0 Runtime (included in self-contained build)
- `truco-core-ffi.dll` (native Go library)

## Building

### Prerequisites

1. Install Go (to build the native DLL)
2. Install .NET 8.0 SDK

### Build Steps

```bash
# 1. Build the native DLL
GOOS=windows GOARCH=amd64 go build -buildmode=c-shared -o bin/truco-core-ffi.dll ./cmd/truco-core-ffi

# 2. Build the WPF app
cd native/windows-wpf
dotnet publish -c Release -o ../../bin/wpf
```

## Portable Build

For a fully portable single-file executable:

```bash
# Build with embedded DLL
cd native/windows-wpf
dotnet publish -c Release -r win-x64 --self-contained true -p:PublishSingleFile=true -o ../../bin/portable-wpf
```

The resulting `TrucoWPF.exe` can be distributed as a single file.

## Project Structure

- `Models.cs` - Data models for game state
- `Services/TrucoCoreService.cs` - P/Invoke wrapper for native library
- `Services/StringProvider.cs` - Localization (pt-BR, en-US)
- `ViewModels/MainViewModel.cs` - MVVM view model
- `MainWindow.xaml` - UI layout
- `MainWindow.xaml.cs` - Code-behind
