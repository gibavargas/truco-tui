## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-06-28 - Token Comparison Timing Attacks
**Vulnerability:** Comparing sensitive strings like tokens, session IDs, and host admin tokens using standard equality operators (`==` or `!=`) is susceptible to timing attacks, as the comparison short-circuits on the first mismatched character.
**Learning:** For any sensitive information that serves as a password, token, or cryptographic key, standard string comparisons can leak the expected value byte by byte over many requests based on response times.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (by converting strings to `[]byte`) to check sensitive tokens against each other, ensuring the comparison time is dependent only on the length of the strings and not on their contents.
