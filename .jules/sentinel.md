## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-05-18 - Hash Map Lookups For Sensitive Tokens
**Vulnerability:** The application stored replacement tokens (sensitive plaintext) directly as keys in a Go map (`h.replaceInvites`) in `internal/netp2p/host.go`. This exposes the tokens in memory and makes map lookups vulnerable to timing attacks.
**Learning:** Storing sensitive plaintext tokens as map keys is insecure. It leaves secrets lingering in memory and creates defense-in-depth vulnerabilities against timing attacks.
**Prevention:** Always hash sensitive tokens (e.g., using `crypto/sha256`) before inserting or looking them up in maps. Store the hex-encoded string of the hash instead of the plaintext token.
