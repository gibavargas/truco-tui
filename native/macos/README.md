# Truco - macOS Native Client (SwiftUI)

A complete, 100% native macOS client for the Go Truco game engine, built with **SwiftUI**. It utilizes a C-ABI FFI wrapper (`libtruco_core.dylib`) to interact with the backend game logic.

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
   Open the Xcode workspace:
   ```bash
   open native/macos/Truco/Truco.xcodeproj
   ```

3. **Build and Run**
   In Xcode, select your Mac as the destination and click the "Run" button (or press `Cmd + R`). The app will launch natively.
