# Truco - macOS Native Client

The single source of truth for the native macOS client is the Xcode app in `native/macos/Truco/Truco`.

This client is built with SwiftUI and uses the shared Go runtime through the C-ABI FFI wrapper `libtruco_core.dylib`.

## Prerequisites
- macOS 13 (Ventura) or newer.
- Xcode 15 or newer.
- Go 1.22+ (for compiling the FFI core).

## Installation & Build Instructions

1. **Build the Go FFI Core**
   From the main project directory, run:
   ```bash
   make ffi-macos
   ```
   This will generate `bin/libtruco_core.dylib` and copy the necessary `truco_core.h` to the Xcode project directory.

2. **Open the Xcode Project**
   Open the Xcode project:
   ```bash
   open native/macos/Truco/Truco.xcodeproj
   ```

3. **Build and Run**
   In Xcode, select your Mac as the destination and click the "Run" button (or press `Cmd + R`). The app will launch natively.

## Notes

- There is no secondary SwiftUI scaffold under `native/macos/`; the Xcode project above is the only macOS implementation that should be edited.
- The FFI build step copies the generated header directly into `native/macos/Truco/Truco/truco_core.h`.
