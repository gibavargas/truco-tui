# Native Desktop Clients

This directory contains native desktop shells that integrate with the shared Go runtime exposed by `cmd/truco-core-ffi`.

## Clients

- `macos`: primary macOS client built with SwiftUI and Xcode
- `linux-gtk`: Linux desktop client built with Rust, GTK4, and libadwaita
- `windows-winui`: primary Windows client built with WinUI 3 and .NET 8
- `windows-wpf`: legacy Windows client kept for compatibility and experimentation

## Architecture

Each shell is intentionally thin:

- the Go runtime owns rules, lobby flow, network orchestration, and snapshots
- the native layer owns rendering, input, accessibility, and platform integration
- the C ABI is the supported boundary between them

## FFI Outputs

- macOS: `bin/libtruco_core.dylib`
- Linux: `bin/libtruco_core.so`
- Windows: `bin/truco-core-ffi.dll`

See the client-specific READMEs in each subdirectory for build steps and platform requirements.
