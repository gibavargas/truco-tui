# Native Desktop Clients

This folder contains the native desktop shells that consume the shared Go core
from `cmd/truco-core-ffi`.

- `macos`: SwiftUI app scaffold for macOS.
- `linux-gtk`: Rust + GTK4/libadwaita scaffold for Linux.
- `windows-winui`: WinUI 3 + MVVM Toolkit scaffold for Windows.

Each client is intentionally thin:

- the Go runtime owns game rules, online orchestration, and failover behavior
- the native shell owns rendering, accessibility, and platform UX
- the C ABI is the only supported integration boundary

The expected FFI build artifacts are:

- macOS: `libtruco_core.dylib` + generated header
- Linux: `libtruco_core.so` + generated header
- Windows: `truco_core.dll` + generated header/import library
