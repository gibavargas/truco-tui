# Client Parity Contract

The shared runtime in `internal/appcore` is the source of truth for every non-TUI client, and the TUI is the reference implementation for behavior.

## Supported intents

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

## Runtime compatibility matrix

- `core_api_version`: snapshot + event contract version exposed by `internal/appcore`
- `protocol_version`: online P2P wire protocol version exposed by `internal/netp2p`
- `snapshot_schema_version`: JSON snapshot schema version for all GUI/browser clients
- Supported locales: `pt-BR`, `en-US`
- Invitation compatibility depends on matching `protocol_version` and valid invite encoding
- GUI/browser clients must tolerate additive JSON fields and prefer runtime-owned derived state over local guesses

## Mode to screen mapping

- `idle`: setup screen
- `host_lobby`, `client_lobby`: lobby screen
- `offline_match`, `host_match`, `client_match`: game screen

## Lobby parity rules

- Show invite key when the local runtime is host.
- Show slots, assigned seat, host seat, and connection state.
- Show slot status using the runtime-derived state contract: `empty`, `occupied_online`, `occupied_offline`, `provisional_cpu`.
- Allow chat in lobby.
- Allow host vote for occupied eligible seats.
- Allow replacement invite actions only when the runtime allows them.
- Show start-match only for host lobby mode.

### Lobby slot state semantics

- `empty`: no player assigned
- `occupied_online`: assigned and connected
- `occupied_offline`: assigned and disconnected
- `provisional_cpu`: disconnected slot currently backed by provisional CPU during a started match
- `can_vote_host`: runtime says the local player may vote for that seat
- `can_request_replacement`: runtime says the local player may create/request a replacement invite for that seat

## Match parity rules

- Render 2-player and 4-player layouts with local player at the bottom.
- Keep online chat/events visible during online matches.
- Use runtime state for action availability:
  - `play` only on local turn with no pending raise
  - `truco` asks or raises according to runtime state
  - `accept`/`refuse` only when the local team is responding
- Use the runtime-derived `ui.actions` contract instead of recomputing seat/team logic locally whenever available.

## Required surfaces

- Setup: player name, player count, locale, offline start, online host, online join
- Online setup parity:
  - Host setup must expose player count and optional `relay_url`.
  - Join setup must expose invite key and desired role (`auto`, `partner`, `opponent`).
  - Invite presentation must allow easy copy/share for both direct and relay invites.
- Lobby: invite key, slot list with badges, chat, start/leave, vote host, replacement invite
- Lobby diagnostics: show runtime-backed connection status, online/offline state, role when present, and event backlog / last error when available from `bundle.connection` or `bundle.diagnostics`
- Match: score, stake ladder, turn indicator, vira, manilha, played cards, own hand, action bar, persistent online chat/events, recent match log
- Match online parity:
  - Keep vote-host and replacement-invite actions reachable during online matches when runtime allows them.
  - Keep chat input reachable during online matches.
- End state:
  - Offline match: replay/new match plus return path.
  - Online match: session close/leave path instead of forcing offline replay.

## Event categories

- `chat`
- `system`
- `replacement_invite`
- `locale_changed`
- `match_updated`
- `lobby_updated`
- `error`

Clients may style these differently, but must preserve their meaning and visibility.

## Locale support

- `pt-BR`
- `en-US`
