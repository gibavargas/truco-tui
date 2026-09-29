## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-09-29 - Timing Attacks in Token Comparisons
**Vulnerability:** Several endpoints in `cmd/truco-relay/main.go` and `internal/netp2p/host.go` used standard string equality operators (`==` and `!=`) to compare sensitive authentication tokens and session credentials.
**Learning:** Standard string comparisons evaluate character by character and return immediately upon the first mismatch. In security contexts, this allows attackers to measure the time taken to reject an invalid token and guess the correct token character by character (a timing attack).
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` when comparing sensitive strings, tokens, or credentials to ensure the comparison time is dependent only on the length of the strings and not their contents.
