## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-05-24 - Timing Attacks in Token Comparisons
**Vulnerability:** Standard equality operators (`==` or `!=`) were being used to compare sensitive strings like `AdminToken` and `joinMsg.Token`.
**Learning:** Using standard equality operators for sensitive strings makes the application susceptible to timing attacks, where an attacker can measure the time taken for the comparison to fail and incrementally guess the token byte-by-byte.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 0` for comparing sensitive strings, tokens, and credentials in Go to ensure the comparison time is independent of the input contents.
