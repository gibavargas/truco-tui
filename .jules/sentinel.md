## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-05-24 - Timing Attacks in Authentication

**Vulnerability:** String comparison operations (`==` and `!=`) were used for validating tokens (`HostAdminToken`, `joinMsg.Token`, and `PeerCredential`). String comparisons in Go (like many other languages) short-circuit when a mismatch is found, meaning comparing the first character takes less time than comparing the whole string. An attacker could use this to leak valid tokens character by character.

**Learning:** Any sensitive token, credential, or session ID used for authentication should be considered private material, and therefore vulnerable to timing side-channels during validation.

**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (by casting the strings to `[]byte` and verifying the result is `1`) whenever comparing sensitive tokens, secrets, or keys instead of relying on the equality operators.
