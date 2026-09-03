## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-04-14 - Timing Attacks in Token Verification
**Vulnerability:** Standard string equality operators (`==` and `!=`) were used to verify security-critical tokens (e.g., `req.HostAdminToken != sess.AdminToken` and `joinMsg.Token != h.token`).
**Learning:** Standard string comparisons short-circuit upon finding the first mismatched character. This allows attackers to iteratively guess the token by measuring the response time (timing attack), severely compromising authentication and session integrity.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` when comparing cryptographic secrets, session tokens, or authentication credentials to ensure the comparison time depends only on the length of the slices and not their contents.
