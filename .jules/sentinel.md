## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2025-02-18 - Timing Attacks on Secret Comparisons
**Vulnerability:** String comparisons for sensitive tokens like `AdminToken` and `Credential` in `cmd/truco-relay/main.go` were using standard string equality `!=`.
**Learning:** Standard string comparisons fail early and can be used in timing attacks to guess the secret value byte by byte by measuring the response time.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` when comparing sensitive strings like tokens, passwords, or credentials.
