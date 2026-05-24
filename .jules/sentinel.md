## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-05-24 - Fix Timing Attacks in Token Comparisons
**Vulnerability:** The codebase was using standard equality operators (`!=`) to compare sensitive tokens (`req.HostAdminToken` vs `sess.AdminToken`, and `joinMsg.Token` vs `h.token`).
**Learning:** Standard string comparisons stop at the first mismatching byte, taking variable amounts of time based on how much of the string matches. This exposes the app to timing attacks where an attacker could guess the token byte-by-byte by measuring response times.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (with `[]byte` casting) when comparing sensitive strings like tokens, passwords, or API keys.
