## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-05-15 - Prevent Timing Attacks in Token Comparisons
**Vulnerability:** Timing attack possible during token validation. Standard equality checks (`==` or `!=`) short-circuit, allowing an attacker to guess a token character-by-character based on response times.
**Learning:** This existed because standard string comparison was used for sensitive `Token` and `HostAdminToken` validations in `internal/netp2p/host.go` and `cmd/truco-relay/main.go`.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` when comparing sensitive tokens, secrets, or passwords to ensure execution time is constant regardless of where the mismatch occurs.
