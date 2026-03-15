# macOS Xcode Project

This directory contains the Xcode project for the macOS client.

## Important Paths

- `Truco.xcodeproj`: project file
- `Truco/`: Swift source, assets, bridge layer, and generated header
- `TrucoTests/`: unit tests
- `TrucoUITests/`: UI tests

## Runtime Integration

The app talks to the shared runtime through `TrucoCoreBridge.swift`, which wraps the exported functions from `cmd/truco-core-ffi`.

The project expects:

- `bin/libtruco_core.dylib`
- `native/macos/Truco/Truco/truco_core.h`

Generate or refresh them from the repository root with:

```bash
make ffi-macos
```

## Open in Xcode

```bash
open native/macos/Truco/Truco.xcodeproj
```

Build and run with `My Mac` as the destination.
