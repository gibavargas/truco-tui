# macOS Native Client

The macOS client is a SwiftUI shell backed by the shared Go runtime through `libtruco_core.dylib`.

## Requirements

- macOS 13+
- Xcode 15+
- Go 1.24+

## Build the Shared Library

From the repository root:

```bash
make ffi-macos
```

That target builds `bin/libtruco_core.dylib` and refreshes the generated C header used by the Xcode project.

## Open the Project

```bash
open native/macos/Truco/Truco.xcodeproj
```

The Xcode project embeds `bin/libtruco_core.dylib` into the app bundle during the build.

## Notes

- The application sources live under `native/macos/Truco/Truco/`.
- UI behavior should stay aligned with the runtime contract in `docs/PARITY.md`.
- If the shared library is stale, rebuild it before launching from Xcode.
