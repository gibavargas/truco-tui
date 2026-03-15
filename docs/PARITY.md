# Client Parity Contract

The shared runtime in `internal/appcore` is the source of truth for all non-TUI clients. The TUI remains the reference UI for behavior, but clients should consume runtime state instead of re-implementing rules locally.

## Version Contract

- `core_api_version`: `1`
- `protocol_version`: `2`
- `snapshot_schema_version`: `1`

Clients must reject incompatible protocol invites and should tolerate additive JSON fields in snapshots and events.

## Supported Locales

- `pt-BR`
- `en-US`

## Supported Intents

- `set_locale`
- `new_offline_game`
- `create_host_session`
- `join_session`
- `start_hosted_match`
- `game_action`
- `send_chat`
- `vote_host`
- `request_replacement_invite`
- `close_session`

## Snapshot Model

Every client should treat `SnapshotBundle` as the rendering contract:

- `versions`: runtime compatibility metadata
- `mode`: current screen and flow state
- `locale`: active locale
- `match`: full match snapshot when a game exists
- `lobby`: online session snapshot when in a session
- `ui`: runtime-derived action and lobby affordance state
- `connection`: online status, host role, last error, and last consumed event sequence
- `diagnostics`: event backlog and replay diagnostics

## Mode Mapping

- `idle`: setup screen
- `host_lobby`: host lobby
- `client_lobby`: joined lobby
- `offline_match`: offline match screen
- `host_match`: online host match screen
- `client_match`: online client match screen

## Setup Parity Rules

- Allow selecting locale before match or session creation.
- Offline setup must expose player name and player count.
- Host setup must expose player name, player count, and optional `relay_url`.
- Join setup must expose invite key, player name, and desired role.
- Desired roles must support `auto`, `partner`, and `opponent`.

## Lobby Parity Rules

- Show the invite key when the local runtime is hosting.
- Render all seats using runtime-derived lobby slot data.
- Keep chat available while in the lobby.
- Show start match only in `host_lobby`.
- Surface runtime-owned host voting and replacement invite affordances.
- Expose runtime connection state and the latest error or diagnostics when available.

Lobby slot statuses:

- `empty`
- `occupied_online`
- `occupied_offline`
- `provisional_cpu`

Per-seat affordances:

- `can_vote_host`
- `can_request_replacement`

## Match Parity Rules

- Render both 2-player and 4-player layouts with the local player at the bottom.
- Keep chat and recent events visible during online matches.
- Use `ui.actions` to decide which actions are available.
- Do not recompute turn ownership, team response state, or raise eligibility from raw match state when runtime-derived action flags already exist.
- Keep host vote and replacement invite actions reachable in online matches when the runtime allows them.

Action expectations:

- `play`: only when `can_play_card` is true
- `truco`: only when `can_ask_or_raise` is true
- `accept`: only when `can_accept` is true
- `refuse`: only when `can_refuse` is true
- leave or close session: only when `can_close_session` is true

## End-State Behavior

- Offline matches should allow starting another offline match or returning to setup.
- Online matches should preserve session-aware leave and close flows instead of forcing offline replay behavior.

## Event Categories

- `chat`
- `system`
- `replacement_invite`
- `locale_changed`
- `match_updated`
- `lobby_updated`
- `error`

Clients may style categories differently, but must keep them visible and semantically distinct.
