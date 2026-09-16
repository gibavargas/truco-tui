## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-04-14 - Timing Attacks in Credential Validation
**Vulnerability:** Standard equality operators (`!=` and `==`) were used to compare sensitive credentials, API tokens, and session identifiers across multiple files (`cmd/truco-relay/main.go`, `internal/netp2p/host.go`).
**Learning:** Standard string comparisons terminate early on the first mismatched byte, allowing an attacker to deduce the contents of the credential one character at a time by measuring the time taken for the server to reject the request (Timing Attacks).
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` to compare sensitive tokens to ensure comparison times are independent of the input contents and length.
