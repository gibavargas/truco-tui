## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2026-07-12 - Timing attacks in credential comparison
**Vulnerability:** Standard string inequality comparisons (`!=`) were being used for comparing sensitive tokens and credentials (e.g. `req.HostAdminToken`, `sess.AdminToken`, `mem.Credential`). This makes the application vulnerable to timing attacks, where an attacker can figure out the secret string character by character by measuring how long the comparison takes.
**Learning:** `crypto/subtle.ConstantTimeCompare` MUST be used whenever comparing sensitive values like tokens, API keys, or credentials to prevent timing attacks.
**Prevention:** Always cast the sensitive strings to `[]byte` and use `crypto/subtle.ConstantTimeCompare`, checking if the return value is `1` to confirm equality (or `!= 1` for inequality).
