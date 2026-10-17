## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-09-20 - Timing Attacks in Token Comparisons
**Vulnerability:** String equality operators (`!=`) were used to compare sensitive authentication tokens and session credentials in `cmd/truco-relay/main.go`.
**Learning:** Using standard string equality operators for sensitive data exposes the application to timing attacks, where an attacker can determine the length of the matching prefix based on the time the comparison takes.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` when comparing sensitive tokens, passwords, or hashes to ensure the comparison time is independent of the input values.
