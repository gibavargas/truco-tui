# Truco Paulista

**Truco Paulista online grátis** — jogue no navegador ou no terminal. Multiplayer P2P, partida contra CPU, chat em tempo real e animações. Construído em Go com Bubble Tea.

> 🎮 **Jogue agora:** [truco.jvguidi.com.br](https://truco.jvguidi.com.br) — sem cadastro, direto no navegador

Monorepo for a Truco Paulista implementation centered on a shared Go runtime.

The repository contains:

- A terminal client built with Bubble Tea and Lip Gloss.
- A browser client backed by a Go HTTP API.
- A Wails desktop app (`desktop/wails`), the supported desktop direction for Windows and Linux.
- Native desktop shells for macOS (supported) plus deprecated Linux GTK and Windows WinUI clients.
- A C++ CPU strategy engine (`native/ai-engine`) linked into the Go runtime through CGO.
- A relay service for restrictive NAT and firewall scenarios.

The shared runtime in `internal/appcore` owns game rules, online session orchestration, snapshots, and event flow. Every non-TUI client integrates with that runtime directly or through a thin transport layer.

## Repository Layout

- `cmd/truco`: terminal client entrypoint.
- `cmd/truco-core-ffi`: C ABI bridge used by native desktop clients.
- `cmd/truco-relay`: HTTPS control plane + QUIC and TLS-over-TCP data plane relay service.
- `browser-edition/`: browser UI, Go HTTP API, and packaging script.
- `docs/`: parity and binary naming contracts.
- `internal/appcore`: runtime API, snapshots, and events.
- `internal/netp2p`: direct P2P transport and invite protocol.
- `internal/netrelay`: relay control-plane client.
- `internal/netquic`: QUIC helpers.
- `internal/truco`: game rules, deck model, and CPU behavior.
- `internal/ui`: Bubble Tea TUI implementation.
- `native/`: native desktop clients and the C++ AI engine (`native/ai-engine`).
- `desktop/wails`: Wails desktop app (Go adapter + TypeScript UI).
- `scripts/`: artifact validation, icon generation, and the relay TCP smoke test (`scripts/e2e-tcp-smoke`).

## Requirements

- Go 1.24+
- Node.js 20+ for the browser and Wails frontend builds
- A C++17 compiler (`g++` or `clang++`) with `make` and CGO enabled to get the C++ CPU engine (`libtruco_ai`). Builds with `CGO_ENABLED=0` (including cross-compiles) still work and fall back to the pure-Go CPU heuristic (`internal/truco/cpu_nocgo.go`)
- Platform-specific toolchains only when building native clients
- Wails CLI v2 only when building the Wails desktop app

With CGO enabled (the default on host builds), build the C++ AI engine once before any `go run`, `go build`, `go vet`, or `go test`, and again whenever `native/ai-engine` changes:

```bash
make -C native/ai-engine
```

## Main Commands

Run the terminal client:

```bash
go run ./cmd/truco
```

Build the terminal client for the current host:

```bash
make build
```

Build the relay binary for the current host:

```bash
make build-relay
```

Run tests:

```bash
go test ./...
```

Start the relay service:

```bash
make relay
```

Package the browser edition:

```bash
make browser
```

Run the packaged browser edition:

```bash
browser-edition/dist/truco-api
```

Then open `http://127.0.0.1:9090/`.

Validate generated artifacts and browser packaging:

```bash
make verify-artifacts
```

Build the shared macOS FFI library:

```bash
make ffi-macos
```

Build the shared Windows FFI library:

```bash
make ffi-windows
```

Build the Linux GTK client:

```bash
make linux-gtk
```

- Builds `bin/libtruco_core.so`
- Copies the library to `native/linux-gtk/lib/libtruco_core.so`
- Compiles the libadwaita client at `native/linux-gtk/target/release/truco-linux-gtk`

Build the Linux Flatpak bundle:

```bash
make flatpak-linux
```

Build the Wails desktop app (see [desktop/wails/README.md](desktop/wails/README.md)):

```bash
make wails-build            # host platform
make wails-build-linux      # linux/amd64
make wails-build-windows    # windows/amd64
make verify-wails           # typecheck, frontend tests, Go adapter tests, dry-run builds
```

## Binary Naming

All produced binaries follow [docs/BINARY_NAMING.md](docs/BINARY_NAMING.md).
Release checks live in [docs/RELEASE_CHECKLIST.md](docs/RELEASE_CHECKLIST.md).

## Multiplayer Overview

- Offline matches support 2 or 4 players with CPU seats.
- Online play supports direct P2P and relay-backed sessions.
- Lobby features include chat, host transfer voting, and replacement invites.
- Clients reconnect automatically after transient network failures.
- If the host process disappears, the session can elect a new host and continue.
- Relay invites use transport version 2 and short-lived join tickets.

## Relay Service

The relay is intended for players behind restrictive NAT or firewall setups.

Run locally:

```bash
go run ./cmd/truco-relay
```

Environment variables:

- `TRUCO_RELAY_HTTP_ADDR`: control-plane HTTPS address. Default: `127.0.0.1:9443`
- `TRUCO_RELAY_QUIC_ADDR`: QUIC tunnel address. Default: `127.0.0.1:9444`
- `TRUCO_RELAY_TCP_ADDR`: TLS-over-TCP tunnel address, used as fallback when UDP is blocked. Default: `:9445`
- `TRUCO_RELAY_TLS_CERT_FILE`: TLS certificate path for production
- `TRUCO_RELAY_TLS_KEY_FILE`: TLS key path for production

Relay protocol notes:

- Online transport uses protocol version 2.
- Relay invites use short-lived `relay_join_ticket` values.
- Clients validate TLS 1.3 and may enforce SPKI pinning.
- The data plane dials QUIC first and falls back to TLS-over-TCP per connection; host and peer may use different transports in the same session. Relay responses advertise `tcp_addr` (additive; older relays omit it and clients derive `host:9445`).
- Smoke-test the TCP path against a running relay with `RELAY_URL=... RELAY_TCP_ADDR=... go run ./scripts/e2e-tcp-smoke`.
- Expired relay sessions, members, and tickets are cleaned up automatically.

Operational endpoints:

- `GET /healthz`
- `GET /metrics`

## Rules Implemented

- 40-card Truco deck without 8, 9, or 10
- Dynamic manilha from the vira
- Paulista ranking order
- Suit order for manilha resolution
- Stake ladder `1 -> 3 -> 6 -> 9 -> 12`
- Match target of 12 points

## Controls in the TUI Match

- `1`, `2`, `3`: play a card
- `t`: ask or raise truco
- `a`: accept
- `r`: refuse
- `tab`: cycle `mesa`, `chat`, and `log`
- `enter`: send chat while focused on the chat tab
- `q`: leave the match

Online chat commands:

- `/host <slot>`: vote to transfer host ownership
- `/invite <slot>`: mint a replacement invite for a disconnected provisional CPU seat

## Documentation Index

- `docs/PARITY.md`: behavior contract for browser and native clients
- `docs/BINARY_NAMING.md`: artifact naming and output locations
- `docs/RELEASE_CHECKLIST.md`: pre-release validation and smoke checks
- `browser-edition/README.md`: browser client workflow
- `desktop/wails/README.md`: Wails desktop app workflow
- `native/README.md`: native client overview
- `progress.md`: current project status

## FAQ

### Como jogar Truco Paulista online?

Acesse [truco.jvguidi.com.br](https://truco.jvguidi.com.br) no navegador. Escolha jogar contra a CPU ou crie uma sala multiplayer e compartilhe o link com amigos. Não precisa de cadastro.

### Preciso instalar algo para jogar?

Não. O jogo roda direto no navegador. Para a versão terminal, instale Go 1.24+ e um compilador C++, rode `make -C native/ai-engine` e depois `go run ./cmd/truco`.

### O Truco Paulista tem multiplayer?

Sim. O jogo suporta multiplayer P2P (ponto a ponto) direto entre jogadores, e também oferece um relay service para contornar NAT restritivo ou firewalls.

### Quais as regras do Truco Paulista implementadas?

Baralho de 40 cartas (sem 8, 9 e 10), manilha dinâmica do vira, ranking Paulista, escala de apostas 1→3→6→9→12, e meta de 12 pontos para vencer a partida.

### Tem versão para desktop?

Sim. Há o app desktop em Wails (Windows e Linux) e o cliente nativo para macOS; os clientes Linux GTK e Windows WinUI estão descontinuados. Veja a seção [Main Commands](#main-commands) para instruções de build de cada plataforma.
