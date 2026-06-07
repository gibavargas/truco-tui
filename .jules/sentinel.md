## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-06-07 - Timing Attacks in Token Verification
**Vulnerability:** String comparisons for sensitive tokens and credentials used the standard `!=` operator, which can leak timing information based on how long the comparison takes (as it fails on the first non-matching character).
**Learning:** This could allow an attacker to bypass authentication by observing the time taken to process requests and slowly guessing tokens character by character.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` when comparing tokens or credentials.
