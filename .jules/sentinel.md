## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-05-18 - Timing Attacks on Token Validation
**Vulnerability:** Token comparison logic in `cmd/truco-relay/main.go` and `internal/netp2p/host.go` used standard inequality operators (`!=`). This allows timing attacks because standard string comparisons abort on the first non-matching byte, leaking information about the token character by character.
**Learning:** Comparing sensitive strings (like authentication tokens or session IDs) securely requires constant-time comparison to prevent side-channel leaks based on execution time.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (by casting strings to `[]byte` and verifying against `1`) when comparing secrets to ensure execution time is uniform regardless of byte matches.
