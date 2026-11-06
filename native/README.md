# Native Desktop Clients

This folder contains the native desktop shells that consume the shared Go core
from `cmd/truco-core-ffi`.

Supported variants:

- `macos`: SwiftUI app for macOS.

Deprecated variants (superseded by the Wails desktop app in `desktop/wails`):

- `linux-gtk`: Rust + GTK4/libadwaita legacy shell for Linux.
- `windows-winui`: WinUI 3 + MVVM Toolkit legacy shell for Windows.

Shared native code:

- `ai-engine`: C++17 CPU strategy engine (`libtruco_ai`). It is linked into the Go runtime via CGO (`internal/truco/cpu_cpp.go`), so build it with `make -C native/ai-engine` before building or testing any CGO-enabled Go target. Builds with `CGO_ENABLED=0` use the pure-Go fallback in `internal/truco/cpu_nocgo.go`.

Each client is intentionally thin:

- the Go runtime owns game rules, online orchestration, and failover behavior
- the native shell owns rendering, accessibility, and platform UX
- the C ABI is the only supported integration boundary

The expected FFI build artifacts are:

- macOS: `libtruco_core.dylib` + generated header
- Linux: `libtruco_core.so` + generated header
- Windows: `truco_core.dll` + generated header/import library
