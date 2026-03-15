# Linux GTK Client

The Linux desktop client uses Rust with GTK4 and libadwaita, while game logic and online behavior come from the shared Go runtime loaded dynamically at startup through `libtruco_core.so`.

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

## Build the GTK App

```bash
make linux-gtk
```

That will:

- build `bin/libtruco_core.so`
- copy it to `native/linux-gtk/lib/libtruco_core.so`
- compile the GTK client in release mode

For development:

```bash
mkdir -p native/linux-gtk/lib
cp bin/libtruco_core.so native/linux-gtk/lib/libtruco_core.so
cargo run --manifest-path native/linux-gtk/Cargo.toml
```

The Linux loader searches for the runtime in this order:

- `TRUCO_CORE_LIB`
- `bin/libtruco_core.so`
- `native/linux-gtk/lib/libtruco_core.so`
- `lib/libtruco_core.so`
- `libtruco_core.so`
- next to the packaged executable

## Flatpak

Manifest:

```bash
native/linux-gtk/dev.truco.Native.yaml
```

Build locally:

```bash
make flatpak-linux
```

## Notes

- The UI entrypoint is `src/main.rs`.
- The layout definition lives in `window.ui`.
- Runtime parity expectations are defined in `docs/PARITY.md`.
