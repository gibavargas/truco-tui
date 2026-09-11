## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-04-14 - Timing Attacks in Authentication Token Comparisons
**Vulnerability:** Several HTTP handlers in `cmd/truco-relay/main.go` used standard string inequality operators (`!=`) to compare authentication tokens (`HostAdminToken` and `PeerCredential`) against expected secrets.
**Learning:** Standard string comparisons in Go terminate early as soon as a mismatch is found. In authentication scenarios, this creates a timing vulnerability where an attacker could theoretically measure the response time to guess the secret character by character, compromising the security of the relay service.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` for equality (and `!= 1` for inequality) when verifying secrets, tokens, or credentials to prevent timing attacks.
