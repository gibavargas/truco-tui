## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-06-10 - Timing Attack Vulnerability in Token Comparisons
**Vulnerability:** Admin and session tokens were being compared using standard string equality (`!=`), which is vulnerable to timing attacks as it fails fast on the first non-matching character.
**Learning:** Even internal or P2P networking protocols need constant-time comparison for sensitive tokens.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (by casting strings to `[]byte` and checking `!= 1`) when verifying tokens, passwords, or secret keys.
