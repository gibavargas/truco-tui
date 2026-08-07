## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-04-14 - Timing Attacks in Token/Credential Comparisons
**Vulnerability:** Several functions in `cmd/truco-relay/main.go` and `internal/netp2p/host.go` used standard string equality operators (`==` or `!=`) to compare sensitive tokens and credentials, which is vulnerable to timing attacks.
**Learning:** Attackers can measure the time taken to compare strings and guess the tokens byte-by-byte. This applies to authentication tokens, credentials, and API keys.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (casting strings to `[]byte`) for sensitive comparisons.
