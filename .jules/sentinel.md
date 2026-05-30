## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2026-05-30 - Prevent Timing Attacks in Token Verification
**Vulnerability:** String comparisons for sensitive values like `HostAdminToken` and `joinMsg.Token` used the `==` or `!=` operators. This normal equality check fails immediately upon the first mismatched byte, creating a timing side-channel that could allow an attacker to reconstruct the token by carefully measuring response times.
**Learning:** Standard string equality comparisons are not suitable for sensitive security parameters because their variable execution time exposes internal state.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` for verifying cryptographic tokens, passwords, session keys, or API tokens to ensure comparisons take a constant amount of time, nullifying timing attacks.
