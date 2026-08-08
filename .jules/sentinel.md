## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-08-08 - Timing Attack on Authentication Tokens
**Vulnerability:** Comparing sensitive authentication tokens using standard string inequality operators (`!=`).
**Learning:** String comparisons short-circuit upon finding a mismatched character. This could allow an attacker to infer the correct token byte-by-byte by measuring response times (timing attack).
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` when verifying secrets to ensure comparison time depends only on the length, not the contents.
