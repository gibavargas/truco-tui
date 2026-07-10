## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-07-10 - Timing Attacks in Token Comparisons
**Vulnerability:** Sensitive strings like access tokens were being compared using standard string equality (`==` or `!=`).
**Learning:** Standard string equality operators short-circuit, which allows timing attacks to guess sensitive strings byte-by-byte by observing the response time.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (by first casting the strings to `[]byte`) when checking tokens, passwords, keys, or hashes.
