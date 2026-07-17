## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-03-11 - Timing attack in token verification
**Vulnerability:** A timing attack vulnerability was identified in `internal/netp2p/host.go` during host session token verification. Sensitive token strings were being compared using standard equality operators (`!=`), which short-circuit and leak token structure information via execution time.
**Learning:** Even internal p2p network tokens are susceptible to timing attacks. Relying on simple string equality (`==` or `!=`) for sensitive authentication parameters exposes the application to enumeration risks.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1` when comparing sensitive strings such as tokens, passwords, or API keys to enforce strict time-constant comparison and avoid information leakage.
