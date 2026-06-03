## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-06-03 - Timing Attacks in Token Comparisons
**Vulnerability:** Token comparisons (like invite tokens or host admin tokens) in `cmd/truco-relay/main.go` and `internal/netp2p/host.go` were using standard string equality `!=` instead of constant-time comparisons.
**Learning:** Using standard string equality on secrets creates a side-channel timing attack where an attacker can guess the token byte-by-byte based on the time it takes the comparison to fail.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` when comparing sensitive tokens, secrets, or passwords.
