## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-06-06 - Use ConstantTimeCompare for Security Tokens
**Vulnerability:** Timing attack vulnerability in token comparisons (req.HostAdminToken vs sess.AdminToken, and joinMsg.Token vs h.token).
**Learning:** Standard string equality operators (`==` or `!=`) short-circuit, leaking information about the token character-by-character through execution time differences, allowing attackers to incrementally guess valid tokens.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` when comparing security tokens, session keys, or credentials.
