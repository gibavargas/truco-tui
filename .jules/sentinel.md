## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2025-05-18 - Timing Attacks on Tokens and Credentials
**Vulnerability:** Several functions in `cmd/truco-relay/main.go` used standard string comparison operators (`!=`) for checking authentication tokens and credentials.
**Learning:** Using standard string comparison operators on sensitive data like tokens and credentials is a critical vulnerability that allows attackers to perform timing attacks to guess the secret value byte-by-byte.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` for equality checks and `!= 1` (or `== 0`) for inequality checks when comparing sensitive cryptographic data, tokens, and credentials.
