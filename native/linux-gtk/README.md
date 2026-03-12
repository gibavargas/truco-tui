# Truco - Linux GTK Native Client

Este é o cliente nativo de Linux para o jogo Truco, construído usando **Rust**, **GTK 4** e **libadwaita**. A lógica do jogo é mantida num back-end compartilhado em **Go** via FFI (Foreign Function Interface).

## Requisitos
- **Linux** (Qualquer distribuição moderna, como Ubuntu, Fedora, Arch, etc).
- **Go** 1.22+ (para a engine do core do jogo).
- **Rust** 1.75+ e o Cargo (via rustup).
- Bibliotecas do sistema para **GTK4** e **libadwaita**.

### Ubuntu/Debian
```bash
sudo apt update
sudo apt install libgtk-4-dev libadwaita-1-dev
```

### Fedora
```bash
sudo dnf install gtk4-devel libadwaita-devel
```

### Arch Linux
```bash
sudo pacman -S gtk4 libadwaita
```

## Compilando e Rodando

Para compilar, precisamos antes gerar e embutir o backend core em Go (`libtrucocore_ffi.so`). O script `build.rs` deve cuidar de apontar para o diretório se ele já estiver mapeado no path, mas você pode compilar a .so do core primeiro:

### 1- Build da Engine
```bash
cd internal/appcore/ffi
go build -buildmode=c-shared -o ../../../native/linux-gtk/libtrucocore_ffi.so trucocore.go
```

### 2- Build do App Rust
```bash
cd native/linux-gtk
cargo build --release
```

Ou execute diretamente para teste de desenvolvimento:
```bash
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$(pwd)
cargo run
```

## Funcionalidades
- Temas GTK Nativo com UI polida e Adwaita (inclui compatibilidade a modo Escuro).
- Jogo Offline contra a CPU.
- Multiplayer Online: Sistema de Salas com códigos de convite.
- Chat embutido da sala.
- Controles de Manilha e Vira renderizados em tempo-real.
