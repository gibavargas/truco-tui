## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-05-26 - Timing Attacks in Token/Credential Verification
**Vulnerability:** Used standard string comparison (`!=`) for verifying sensitive tokens and credentials (`AdminToken`, `Credential`, `Token`), leaving the application vulnerable to timing attacks where an attacker could deduce tokens byte-by-byte.
**Learning:** Using standard string comparison operators for cryptographic secrets or authentication tokens creates a timing side-channel. The time taken to reject an invalid token varies depending on how many leading bytes match the expected token.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` when comparing sensitive strings like tokens, passwords, or credentials. Remember to convert strings to byte slices and check against `1` (true).
