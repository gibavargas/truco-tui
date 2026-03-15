# Linux GTK Client

The Linux desktop client uses Rust with GTK4 and libadwaita, while game logic and online behavior come from the shared Go runtime loaded dynamically at startup.

## Requirements

- A modern Linux distribution
- Go 1.24+
- Rust toolchain with Cargo
- GTK4 and libadwaita development packages

Ubuntu or Debian:

```bash
sudo apt update
sudo apt install libgtk-4-dev libadwaita-1-dev
```

Fedora:

```bash
sudo dnf install gtk4-devel libadwaita-devel
```

Arch Linux:

```bash
sudo pacman -S gtk4 libadwaita
```

## Build the Shared Library

From the repository root:

```bash
make ffi-linux
```

This produces `bin/libtruco_core.so`.

The current Rust loader expects the library to be available as `libtruco-core-ffi.so`, so either provide a matching filename or adjust the loader before packaging.

## Build the GTK App

```bash
cd native/linux-gtk
cargo build --release
```

For development:

```bash
cd native/linux-gtk
cargo run
```

## Notes

- The UI entrypoint is `src/main.rs`.
- The layout definition lives in `window.ui`.
- Runtime parity expectations are defined in `docs/PARITY.md`.
