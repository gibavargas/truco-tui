## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-07-13 - [Prevent Timing Attacks on Security Tokens and Credentials]
**Vulnerability:** String comparison operators (`==` and `!=`) were used directly to compare sensitive security values (e.g. HostAdminToken, Credentials, JoinToken) in `cmd/truco-relay/main.go` and `internal/netp2p/host.go`. This opens the application up to timing attacks where an attacker can determine a secret byte-by-byte by observing response times.
**Learning:** In Go, string comparisons short-circuit as soon as a differing character is found, leaking length and prefix matching information via timing differences.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (after casting strings to `[]byte` and checking `== 1`) to compare security-sensitive tokens, passwords, and API keys.
