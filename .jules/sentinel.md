## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-07-22 - Fix Timing Attack Vulnerability in Credential Comparison
**Vulnerability:** Found sensitive tokens (`AdminToken`, `Credential`) and P2P join tokens being compared using standard string equality (`==` and `!=`).
**Learning:** Simple string comparison operators in Go can exit early on the first mismatched character, allowing an attacker to determine the valid token byte-by-byte by analyzing the response timing.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` when comparing any sensitive secrets, passwords, tokens, or credentials in Go.
