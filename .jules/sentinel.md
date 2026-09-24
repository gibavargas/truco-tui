## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-04-14 - Timing attacks on Token Comparisons
**Vulnerability:** The application was using standard string inequality operators (`!=`) to compare sensitive authentication tokens (`HostAdminToken` in `cmd/truco-relay/main.go` and `Token` in `internal/netp2p/host.go`).
**Learning:** Comparing sensitive strings like tokens with `!=` or `==` exposes the application to timing attacks, as the standard comparison returns immediately upon finding the first mismatching byte. This allows attackers to infer valid tokens byte-by-byte through side channels.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` for equality and `!= 1` for inequality when comparing sensitive cryptographic data, session IDs, passwords, or authentication tokens.
