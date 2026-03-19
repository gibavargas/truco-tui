# Client Parity Contract

`internal/appcore` is the authoritative contract for every adapter layer in this repository. Browser and native clients must render `SnapshotBundle`, dispatch `AppIntent`, and react to `AppEvent` without re-implementing game rules, lobby policy, turn ownership, or online session behavior locally.

## Runtime Contract

- `core_api_version`: `1`
- `protocol_version`: `2`
- `snapshot_schema_version`: `2`
- Supported locales: `pt-BR`, `en-US`
- Supported desired roles: `auto`, `partner`, `opponent`
- Supported modes:
  - `idle`
  - `host_lobby`
  - `client_lobby`
  - `offline_match`
  - `host_match`
  - `client_match`
- Supported intents:
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
- Stable event kinds that adapters must understand:
  - `chat`
  - `client_joined`
  - `error`
  - `failover_promoted`
  - `failover_rejoined`
  - `host_created`
  - `lobby_updated`
  - `locale_changed`
  - `match_started`
  - `match_updated`
  - `replacement_invite`
  - `session_closed`
  - `session_ready`
  - `system`

Clients must tolerate additive JSON fields in snapshots and events. Clients must not assume event queues are replayable after they have been drained by the adapter.
Any change to the runtime JSON contract must bump `snapshot_schema_version` and the native adapters' required schema constant in the same change, so stale FFI libraries fail fast instead of rendering partial state.

## Snapshot and Error Rules

- `SnapshotBundle` is the rendering contract. Clients should prefer `bundle.ui.actions` and `bundle.ui.lobby_slots` over recomputing actions or affordances from raw match/lobby data.
- `bundle.connection.last_error` is the canonical runtime error surface for adapter-driven failures.
- The browser HTTP API must return `error_code` alongside `error` for non-OK responses. Native clients should surface the runtime error `code` and `message` from FFI responses.
- `close_session` must always return the runtime to `idle`, clear `match` and `lobby`, and emit `session_closed`.

## User Journeys

### Offline 2-player and 4-player setup

- Locale selection is available before starting the game.
- Offline setup accepts player name and player count.
- Starting an offline match must transition from `idle` to `offline_match`.
- Starting a match emits `match_updated` and `session_ready`.
- The local player remains seat `0`; other seats may be CPU-controlled.

### Host flow

- Host setup accepts player name, player count, optional `relay_url`, and optional transport selection if exposed by the client.
- Starting a host session transitions to `host_lobby`, returns a visible invite key, and emits `host_created` and `lobby_updated`.
- The host can start the match only when runtime preconditions are satisfied.
- Starting the hosted match transitions to `host_match`, emits `match_started`, and updates the match snapshot for all adapters.

### Join flow

- Join setup accepts invite key, player name, and desired role.
- Joining a valid session transitions to `client_lobby` first and later to `client_match` when match state arrives.
- The assigned seat, current host seat, connectivity map, and lobby slots come from the runtime; adapters must not infer them independently.

### Relay flow

- Relay-backed invites must use protocol version `2`.
- Join failures caused by invalid relay URL, bad tickets, expired tickets, auth failures, or incompatible protocol must surface a structured runtime error.
- Relay and direct sessions share the same runtime modes, lobby shape, and event categories.

### Chat, host transfer, and replacement invites

- Chat is available in online lobbies and online matches.
- Offline chat is local-only and still uses the `chat` event category.
- Host transfer uses `vote_host` and runtime-derived `can_vote_host` affordances from `bundle.ui.lobby_slots`.
- Replacement invites use `request_replacement_invite`, must respect host/disconnect preconditions, and emit `replacement_invite` when minted successfully.

### Match actions

- Clients must gate actions using `bundle.ui.actions`:
  - `play` only when `can_play_card` is true
  - `truco` only when `can_ask_or_raise` is true
  - `accept` only when `can_accept` is true
  - `refuse` only when `can_refuse` is true
  - leave/close only when `can_close_session` is true
- Clients must not recompute response state from `PendingRaise*` when `bundle.ui.actions` already answers that question.

### Disconnect, reconnect, and host failover

- Disconnect and reconnect status must surface through runtime events and `bundle.connection.last_error`.
- If the host disappears and failover succeeds:
  - promoted clients emit `failover_promoted` and move to `host_match`
  - rejoined clients emit `failover_rejoined` and return through `client_lobby` before resuming match updates
- If failover cannot proceed, the runtime emits `error` with a concrete code such as `failover_unavailable` or `failover_rejoin_failed`.

### Session close

- Resetting or leaving a session uses `close_session`.
- After `close_session`, all clients must render setup again from `idle` and must not reuse stale lobby or match data.

## Adapter Rules

- Browser HTTP API actions must map one-to-one to runtime intents or runtime snapshot/event retrieval.
- FFI adapters must preserve runtime JSON semantics: create runtime, dispatch intent, snapshot poll, event poll, versions query, and string cleanup.
- TUI remains the product reference for interaction quality, but parity for rules and session behavior is owned by `internal/appcore`, not by TUI-specific code.
