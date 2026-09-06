## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-09-06 - Timing Attacks in Token Comparisons
**Vulnerability:** Standard string equality operators (`!=`, `==`) were used to compare sensitive authentication tokens (like `HostAdminToken`, `joinMsg.Token` and `mem.Credential`). This allows attackers to potentially infer valid tokens byte-by-byte via timing attacks, as standard comparisons short-circuit on the first mismatch.
**Learning:** Any comparison involving sensitive secrets, tokens, or credentials must be done in constant time to prevent timing side-channels, regardless of the transport layer encryption.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` for equality (and `!= 1` or `== 0` for inequality) when comparing secrets.
