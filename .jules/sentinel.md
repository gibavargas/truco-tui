## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-05-24 - Timing Attacks via Standard String Comparisons
**Vulnerability:** Several authentication routines in `cmd/truco-relay/main.go` and `internal/netp2p/host.go` compared sensitive tokens and credentials using the standard `!=` string operator.
**Learning:** Using standard equality operators for sensitive data allows attackers to perform timing attacks, deducing the secret character by character based on how long the comparison takes to fail.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` for equality and `!= 1` for inequality when comparing security-sensitive strings or byte slices in Go.
