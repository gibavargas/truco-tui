## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2026-05-28 - Secure Token Comparisons
**Vulnerability:** Authentication and session tokens were being compared using standard string equality (`==` or `!=`) in `cmd/truco-relay/main.go` and `internal/netp2p/host.go`.
**Learning:** Standard string comparisons fail fast on the first mismatched character, allowing an attacker to determine the exact character at which a guess failed by measuring response times (a timing side-channel attack). This makes token guessing significantly easier.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` when comparing security tokens, session IDs, passwords, or other secrets to ensure the comparison time is independent of the input values.
