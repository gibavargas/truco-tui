## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-05-18 - Timing attacks on tokens and plaintext secrets in maps
**Vulnerability:** Comparing tokens using simple string equality (`==` or `!=`) exposes the system to timing attacks. Storing plaintext authentication or replacement tokens in maps allows them to linger in memory or leak in core dumps.
**Learning:** Security tokens should always be verified in constant time. Tokens stored long-term in memory should be hashed.
**Prevention:** Use `crypto/subtle.ConstantTimeCompare` when comparing sensitive tokens. Hash sensitive tokens (e.g. `sha256.Sum256`) before storing them as map keys.
