## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-05-18 - Timing Attacks on Token Validation using Standard Equality Operators
**Vulnerability:** The relay server (`cmd/truco-relay/main.go`) and P2P host (`internal/netp2p/host.go`) used standard equality operators (`!=`) to compare sensitive admin and session tokens.
**Learning:** Comparing sensitive secrets like authentication or replacement tokens with standard string equality operators allows timing attacks. An attacker can measure the time taken to reject an invalid token and deduce the correct token byte-by-byte because standard comparison returns early upon finding the first mismatch.
**Prevention:** When comparing sensitive strings like authentication tokens or credentials in Go, always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 0` to prevent timing attacks.
