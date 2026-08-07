## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-08-07 - Use ConstantTimeCompare for Sensitive Tokens
**Vulnerability:** String comparison (`==` or `!=`) was being used for sensitive authentication and replacement tokens in `cmd/truco-relay/main.go` and `internal/netp2p/host.go`. This opens up the possibility for timing attacks, where an attacker can deduce the tokens byte by byte based on the time it takes the comparison to fail.
**Learning:** Comparing cryptographic or otherwise sensitive tokens using standard string comparison operators exposes the application to timing attacks. Go's standard string equality checks return early as soon as a mismatch is found.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` when comparing sensitive tokens (such as authentication tokens, passwords, keys, or credentials).
