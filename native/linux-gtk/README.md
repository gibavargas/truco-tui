# Truco - Linux GTK Native Client

Cliente nativo Linux em **Rust + GTK4 + libadwaita**, usando o runtime compartilhado em **Go** via FFI (`libtruco_core.so`).

## Requisitos
- Linux moderno
- Go 1.24+
- Rust/Cargo
- GTK4 + libadwaita

### Ubuntu / Debian
```bash
sudo apt update
sudo apt install libgtk-4-dev libadwaita-1-dev
```

### Fedora
```bash
sudo dnf install gtk4-devel libadwaita-devel
```

### Arch
```bash
sudo pacman -S gtk4 libadwaita
```

## Build local
Do repositório raiz:

```bash
make linux-gtk
```

Isso faz:
- build do core FFI em `bin/libtruco_core.so`
- cópia do `.so` para `native/linux-gtk/lib/libtruco_core.so`
- build release do app Rust

## Rodar em desenvolvimento
```bash
make ffi-linux
mkdir -p native/linux-gtk/lib
cp bin/libtruco_core.so native/linux-gtk/lib/libtruco_core.so
cargo run --manifest-path native/linux-gtk/Cargo.toml
```

O app procura a biblioteca nesta ordem:
- `TRUCO_CORE_LIB`
- `bin/libtruco_core.so`
- `native/linux-gtk/lib/libtruco_core.so`
- `lib/libtruco_core.so`
- `libtruco_core.so`
- ao lado do executável empacotado

## Flatpak
Manifest principal:

```bash
native/linux-gtk/dev.truco.Native.yaml
```

Build local:

```bash
make flatpak-linux
```

## Estado atual
- Offline 2p / 4p
- Host / join online
- Chat, voto de host e convite de substituição
- Banner/toast para erros e eventos
- Verificação de compatibilidade do core via `TrucoCoreVersionsJSON`
