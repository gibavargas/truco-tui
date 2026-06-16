## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-04-14 - Timing Attacks in Token/Credential Comparisons
**Vulnerability:** Authentication and credential validations in `cmd/truco-relay/main.go` and `internal/netp2p/host.go` used standard string equality operators (`==` and `!=`).
**Learning:** Comparing sensitive secrets like authentication tokens or credentials with standard equality operators exposes the application to timing side-channel attacks, allowing an attacker to deduce the expected token character-by-character based on the time the comparison takes to fail.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (by casting strings to `[]byte` and checking if it returns 1) when comparing sensitive strings (like tokens, passwords, or API keys).
