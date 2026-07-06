## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2026-07-06 - Timing Attack Vulnerability in Token Comparisons
**Vulnerability:** String tokens (`HostAdminToken` and `Token`) were being compared using standard string inequality operators (`!=`).
**Learning:** Standard string comparisons in Go exit early upon finding the first non-matching byte, which leaks timing information. An attacker could measure the response time to guess tokens byte by byte (timing attack).
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` when comparing sensitive strings or tokens to prevent timing attacks. Convert the strings to byte slices first: `subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1`.
