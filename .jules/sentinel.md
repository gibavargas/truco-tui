## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-05-24 - Timing Attacks in Token Comparison
**Vulnerability:** Admin and session tokens were being compared using standard string inequality operators (`!=`) in `cmd/truco-relay/main.go` and `internal/netp2p/host.go`.
**Learning:** Standard string comparisons in Go exit early upon finding the first mismatched character. This introduces a timing attack vulnerability, where an attacker could theoretically guess the token character by character based on micro-differences in response times.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b))` when comparing sensitive tokens, secrets, or passwords to ensure comparison time is independent of the string contents.
