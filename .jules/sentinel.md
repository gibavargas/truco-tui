## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-05-18 - Timing Attacks on Secrets via String Comparison
**Vulnerability:** Authentication endpoints used simple string equality (`==` or `!=`) to compare tokens and credentials (like `HostAdminToken` or `Credential`), allowing attackers to potentially recover secrets through timing attacks.
**Learning:** Comparing sensitive secrets byte-by-byte with early exit operators (`==` or `!=`) creates a side-channel vulnerability where the comparison time depends on how many characters matched.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` when comparing sensitive tokens, credentials, or keys in Go.
