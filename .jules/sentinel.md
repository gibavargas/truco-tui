## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-07-20 - [Fix timing attack on session token comparisons]
**Vulnerability:** Comparing session tokens using `!=` and map lookups using plaintext sensitive tokens allows timing attacks.
**Learning:** Standard string comparisons and map lookups leak timing information. In Go, `map` lookups cannot easily be made constant-time if the key is the plaintext token.
**Prevention:** Use `subtle.ConstantTimeCompare` for direct string comparisons, and hash sensitive tokens (e.g. using SHA-256) before storing them as map keys.
