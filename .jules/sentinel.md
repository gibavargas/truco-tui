## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-05-31 - Timing Attacks in Token Comparisons
**Vulnerability:** In `cmd/truco-relay/main.go` and `internal/netp2p/host.go`, sensitive token values were being compared using standard string equality (`!=`).
**Learning:** Using standard string equality for sensitive values like session tokens or admin tokens is vulnerable to timing attacks, as the comparison returns early upon finding a mismatch, potentially leaking the token character by character.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (by casting strings to `[]byte` and checking `== 1`) when comparing sensitive strings (like tokens, passwords, or API keys).
