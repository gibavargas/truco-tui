## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-05-24 - Timing Attacks in Credential and Token Comparisons
**Vulnerability:** String comparisons for authentication tokens and credentials in `cmd/truco-relay/main.go` and `internal/netp2p/host.go` used standard equality operators (`!=`), making them vulnerable to timing attacks. An attacker could potentially measure response times to guess valid tokens character by character.
**Learning:** In Go, string comparison operations (`==` and `!=`) short-circuit on the first mismatched byte. When dealing with sensitive data like passwords, API keys, or session tokens, this behavior exposes side-channel information.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` for equality and `!= 1` for inequality when comparing sensitive strings.
