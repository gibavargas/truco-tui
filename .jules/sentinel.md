## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-07-26 - Timing Attack Vulnerability in Token Comparisons
**Vulnerability:** String comparison operations using `!=` or `==` on sensitive data such as authentication tokens and peer credentials in `cmd/truco-relay/main.go` and `internal/netp2p/host.go` expose the system to timing attacks.
**Learning:** Standard string equality checks short-circuit on the first mismatched character. Attackers can measure the response time to iteratively guess valid tokens character by character.
**Prevention:** Always cast sensitive strings to `[]byte` and use `crypto/subtle.ConstantTimeCompare(a, b) == 1` to perform comparisons that take a consistent amount of time regardless of whether the strings match or not.
