## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-05-18 - Timing Attacks in Token Comparison
**Vulnerability:** The relay server (`cmd/truco-relay/main.go`) and P2P host (`internal/netp2p/host.go`) used standard string equality operators (`==` / `!=`) to validate authentication tokens (HostAdminToken and invite tokens).
**Learning:** Standard string equality comparisons are vulnerable to timing attacks. An attacker can measure the time it takes for the server to reject a token to guess it character-by-character, as the comparison fails faster if the first character is wrong, slightly slower if the second is wrong, etc.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` for equality and `!= 1` for inequality when comparing sensitive strings like authentication tokens, passwords, or hashes to ensure the comparison time depends only on string length.
