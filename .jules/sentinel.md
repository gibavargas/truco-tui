## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-04-14 - Timing Attacks in Token Verification
**Vulnerability:** String comparison operators (`!=` or `==`) were used to verify security tokens (`HostAdminToken` and `joinMsg.Token`). This exposes the application to timing attacks, where an attacker can measure the time taken to reject an invalid token and iteratively guess it byte-by-byte because standard string comparisons return immediately on the first mismatch.
**Learning:** All cryptographic secrets, authentication tokens, and sensitive comparisons in Go must use `crypto/subtle.ConstantTimeCompare` to ensure the comparison time depends only on the length of the strings and not their contents.
**Prevention:** Enforce the use of `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` for all equality checks involving sensitive credentials or tokens.
