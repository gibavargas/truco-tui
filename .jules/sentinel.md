## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-05-24 - Timing Attacks on Authentication Tokens
**Vulnerability:** String equality operators (`==`, `!=`) were used to compare authentication tokens (`Token` in `host.go` and `AdminToken` in `main.go`). This is vulnerable to timing attacks where an attacker can guess the token byte-by-byte based on the time it takes the comparison to fail.
**Learning:** Standard string comparisons terminate early on the first mismatched byte. Sensitive credentials like tokens or keys should never be compared using `==` or `!=`.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (casting strings to `[]byte` and checking `== 1` or `!= 1`) when comparing sensitive strings to prevent timing attacks.
