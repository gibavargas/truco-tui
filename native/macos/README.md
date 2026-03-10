# macOS SwiftUI shell

This scaffold targets a native macOS client with:

- a shared `@MainActor` app store
- split navigation for lobby, table, and diagnostics
- a thin Swift bridge over the Go C ABI

The generated `truco_core.h` header from `go build -buildmode=c-shared` should
be added to the Xcode target or SwiftPM system library target before compiling.
