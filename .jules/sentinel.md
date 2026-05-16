## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2026-05-16 - Prevent Timing Attacks on Credentials
**Vulnerability:** Credentials and tokens were compared using standard string equality (`!=` or `==`), exposing them to timing attacks where an attacker could deduce characters sequentially.
**Learning:** Standard string comparisons terminate early upon finding a mismatch, meaning the time it takes to compare strings correlates with the length of their common prefix. This allows attackers to guess secrets letter-by-letter.
**Prevention:** Always use `subtle.ConstantTimeCompare([]byte(secret1), []byte(secret2))` from the `crypto/subtle` package when comparing passwords, API keys, tokens, or any other sensitive material.
