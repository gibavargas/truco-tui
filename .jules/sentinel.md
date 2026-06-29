## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2026-06-29 - [Timing Attack in Authentication Token Comparison]
**Vulnerability:** The application was using the standard equality operator (`!=`) to verify security authentication tokens (`joinMsg.Token != h.token`) during peer connection handshakes.
**Learning:** Standard string comparisons fail fast. They compare character-by-character and return as soon as a mismatch occurs. This creates a measurable difference in execution time, which attackers can exploit to deduce tokens by making many queries.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (after casting strings to `[]byte` and checking `== 1`) when comparing secrets, passwords, tokens, or HMACs.
