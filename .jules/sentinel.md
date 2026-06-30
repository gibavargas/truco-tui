## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2025-06-30 - Prevent Timing Attacks in Token Comparisons
**Vulnerability:** The relay server (`cmd/truco-relay/main.go`) and P2P host (`internal/netp2p/host.go`) compared sensitive tokens (like `HostAdminToken` and `joinMsg.Token`) using standard string equality `==` or `!=`. This allowed attackers to use timing attacks to guess the tokens byte-by-byte by observing the response time.
**Learning:** Comparing cryptographic tokens or secrets using standard comparison operators is unsafe because the comparison stops at the first mismatched byte, revealing how many bytes were correct based on timing.
**Prevention:** Always cast strings to byte slices and use `crypto/subtle.ConstantTimeCompare(a, b) == 1` when comparing sensitive tokens, secrets, or passwords to ensure execution time is independent of the inputs' contents.
