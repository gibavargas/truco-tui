## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-05-24 - Fix Timing Attacks in Token Comparisons
**Vulnerability:** Timing attack on sensitive string comparison (tokens, credentials) by using standard equality operators (`==` or `!=`).
**Learning:** Standard operators short-circuit, so comparing strings takes less time if the first characters don't match compared to if they match but a later character doesn't. An attacker can repeatedly guess a token character by character by measuring response time.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` instead of `==` or `!=` for sensitive tokens, passwords, API keys, or any secret data.
