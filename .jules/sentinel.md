## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-09-07 - Timing Attacks on Authentication Tokens
**Vulnerability:** Standard string inequality comparisons (`!=`) were used to compare sensitive authentication tokens (`joinMsg.Token != h.token` and `req.HostAdminToken != sess.AdminToken`), which is susceptible to timing attacks.
**Learning:** Standard string comparisons terminate early upon finding the first mismatching character, leaking information about the expected token length and content through the execution time.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` when comparing sensitive strings like tokens, passwords or secrets to prevent timing attacks.
