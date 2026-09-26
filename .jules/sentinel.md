## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-04-14 - Timing attacks in token comparison
**Vulnerability:** standard string equality `!=` operators were used for security token comparison in `cmd/truco-relay/main.go` and `internal/netp2p/host.go` which can lead to timing attacks.
**Learning:** Regular string comparison operators evaluate character by character and return as soon as a difference is found, taking a variable amount of time. An attacker can exploit this variable time delay to determine when the server rejects a token and iteratively discover the token string.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` for equality and `!= 1` for inequality to prevent timing attacks when comparing sensitive strings like authentication tokens or credentials in Go.
