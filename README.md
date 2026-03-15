# Truco Paulista

Monorepo for a Truco Paulista implementation centered on a shared Go runtime.

The repository contains:

- A terminal client built with Bubble Tea and Lip Gloss.
- A browser client backed by a Go HTTP API.
- Native desktop shells for macOS, Linux GTK, Windows WinUI, and a legacy Windows WPF client.
- A relay service for restrictive NAT and firewall scenarios.

The shared runtime in `internal/appcore` owns game rules, online session orchestration, snapshots, and event flow. Every non-TUI client integrates with that runtime directly or through a thin transport layer.

## Repository Layout

- `cmd/truco`: terminal client entrypoint.
- `cmd/truco-core-ffi`: C ABI bridge used by native desktop clients.
- `cmd/truco-relay`: HTTPS + QUIC relay service.
- `browser-edition/`: browser UI, Go HTTP API, and packaging script.
- `docs/`: parity and binary naming contracts.
- `internal/appcore`: runtime API, snapshots, and events.
- `internal/netp2p`: direct P2P transport and invite protocol.
- `internal/netrelay`: relay control-plane client.
- `internal/netquic`: QUIC helpers.
- `internal/truco`: game rules, deck model, and CPU behavior.
- `internal/ui`: Bubble Tea TUI implementation.
- `native/`: native desktop clients.

## Requirements

- Go 1.24+
- PHP 8.1+ for the browser edition frontend
- Platform-specific toolchains only when building native clients

## Main Commands

Run the terminal client:

```bash
go run ./cmd/truco
```

Build the terminal client for the current host:

```bash
make build
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

Build the shared macOS FFI library:

```bash
make ffi-macos
```

Build the shared Windows FFI library:

```bash
make ffi-windows
```

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
- `TRUCO_RELAY_TLS_CERT_FILE`: TLS certificate path for production
- `TRUCO_RELAY_TLS_KEY_FILE`: TLS key path for production

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
- `browser-edition/README.md`: browser client workflow
- `native/README.md`: native client overview
