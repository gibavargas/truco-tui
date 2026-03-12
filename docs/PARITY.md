# Client Parity Contract

The shared runtime in `internal/appcore` is the source of truth for every non-TUI client.

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

## Mode to screen mapping

- `idle`: setup screen
- `host_lobby`, `client_lobby`: lobby screen
- `offline_match`, `host_match`, `client_match`: game screen

## Lobby parity rules

- Show invite key when the local runtime is host.
- Show slots, assigned seat, host seat, and connection state.
- Allow chat in lobby.
- Allow host vote for occupied eligible seats.
- Allow replacement invite actions only when the runtime allows them.
- Show start-match only for host lobby mode.

## Match parity rules

- Render 2-player and 4-player layouts with local player at the bottom.
- Keep online chat/events visible during online matches.
- Use runtime state for action availability:
  - `play` only on local turn with no pending raise
  - `truco` asks or raises according to runtime state
  - `accept`/`refuse` only when the local team is responding

## Locale support

- `pt-BR`
- `en-US`
