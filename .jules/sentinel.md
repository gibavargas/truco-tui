## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-07-14 - Timing Attacks in Token/Key Verification
**Vulnerability:** Found multiple instances where sensitive strings like tokens and keys were compared using standard string equality `!=` or `==` (e.g. `req.HostAdminToken != sess.AdminToken` and `joinMsg.Token != h.token`).
**Learning:** Standard string equality checks evaluate character by character and return as soon as a mismatch is found. This leaks information about how many characters are correct based on the time it takes to process the request, leading to timing attacks that can be exploited to guess sensitive tokens.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(token1), []byte(token2)) == 1` when comparing sensitive tokens, secrets, or passwords to prevent timing attacks.
