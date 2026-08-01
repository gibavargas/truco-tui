## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-08-01 - Fix Timing Attacks on Tokens
**Vulnerability:** String tokens like `HostAdminToken` and `joinMsg.Token` were compared using `!=`, which performs a byte-by-byte comparison and can leak timing information, making it vulnerable to timing attacks.
**Learning:** The application uses plain equality operators for sensitive tokens.
**Prevention:** Use `crypto/subtle.ConstantTimeCompare([]byte(token1), []byte(token2)) == 1` for sensitive string comparisons.
