## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2026-07-09 - [Fix timing attack in token/credential comparisons]
**Vulnerability:** String comparison (== and !=) used for sensitive tokens and credentials
**Learning:** Standard string comparison operators return early on the first mismatched byte, allowing attackers to perform timing attacks to guess the secret byte by byte.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` when comparing sensitive strings (like tokens, passwords, or API keys). Convert strings to byte slices and check against 1.
