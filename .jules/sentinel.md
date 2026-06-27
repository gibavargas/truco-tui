## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-06-27 - Timing Attacks in Token Validation
**Vulnerability:** Token strings in both `internal/netp2p/host.go` (P2P lobby tokens) and `cmd/truco-relay/main.go` (Host Admin Tokens) were being compared using standard string inequality operators (`!=`).
**Learning:** Comparing sensitive strings or tokens with standard equality operators allows attackers to exploit timing side-channels. Because `==` and `!=` return early upon finding the first mismatched byte, an attacker can measure the response time to guess a token character-by-character, potentially bypassing authentication.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (by casting string variables to `[]byte` and checking for `== 1`) when comparing secrets, passwords, tokens, or API keys to ensure the comparison time depends solely on the token length, neutralizing timing attacks.
