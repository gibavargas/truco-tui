## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-06-15 - [Prevent Timing Attacks on Token Comparison]
**Vulnerability:** Comparing sensitive strings like `joinMsg.Token` against `h.token` using standard string inequality (`!=`) in `internal/netp2p/host.go` exposes the application to timing attacks.
**Learning:** Standard string equality operators return as soon as a mismatch is found, allowing an attacker to deduce a token based on the time the comparison takes.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` when comparing sensitive tokens, secrets, passwords, or API keys.
