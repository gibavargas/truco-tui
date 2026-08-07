## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-08-04 - Timing Attack on Relay Server Authentication
**Vulnerability:** The relay server in `cmd/truco-relay/main.go` verified session admin tokens and peer credentials using standard string equality comparisons (e.g., `req.HostAdminToken != sess.AdminToken`). These non-constant time comparisons exposed the authentication endpoints to timing side-channel attacks, potentially allowing an attacker to incrementally guess valid tokens.
**Learning:** Security tokens, passwords, and sensitive API keys must never be compared using `==` or `!=` operators in Go, as these operations short-circuit and execution time depends on the length of the matching prefix.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (after casting strings to `[]byte` and checking for `== 1`) when verifying sensitive authentication tokens to ensure comparison time is independent of the input string's contents.
