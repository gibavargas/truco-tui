## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2026-06-12 - Prevent Timing Attacks on Secure Tokens
**Vulnerability:** Tokens and session secrets were compared using standard string equality (`==` or `!=`), making them susceptible to timing side-channels.
**Learning:** A standard equality operator halts at the first mismatch. An attacker can theoretically measure the time taken to evaluate the comparison and guess characters one by one, ultimately recovering sensitive tokens.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (casting to `[]byte`) to compare sensitive tokens securely.
## 2026-06-12 - Flaky Test Resolution
**Learning:** The `TestPlayFaceDownCardUsesFaceDownPath` test in `desktop/wails/app_test.go` was flaky because it relied on `StartOfflineGame` which initializes a game with a random seed. This led to unpredictable hands where the first trick might not be playable immediately.
**Prevention:** For tests requiring specific offline game states, explicitly construct and dispatch a `NewOfflineGamePayload` with fixed `SeedLo` and `SeedHi` values to ensure deterministic gameplay behavior.
