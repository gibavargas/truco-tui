## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-04-14 - Timing Attack Vulnerability in Token Comparisons
**Vulnerability:** Several places in the codebase (`cmd/truco-relay/main.go`, `internal/netp2p/host.go`) compared sensitive strings (`HostAdminToken`, `Credential`, `Token`) using the standard `!=` operator, which can leak timing information and allow an attacker to guess the token byte-by-byte.
**Learning:** Comparing sensitive tokens using standard string equality creates a timing channel. The time it takes to compare strings depends on the first character that differs.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (casting strings to `[]byte` and checking `== 1`) instead of standard equality operators (`==` or `!=`) when comparing sensitive strings (like tokens, passwords, or API keys).
