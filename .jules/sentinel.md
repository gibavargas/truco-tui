## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-05-20 - Timing Attacks on Authentication Tokens
**Vulnerability:** Comparing sensitive strings (like `HostAdminToken`, `Credential`, `Token`, `sessionID`) using standard string equality (`==` or `!=`) instead of `crypto/subtle.ConstantTimeCompare`. This allows attackers to perform timing attacks to infer the sensitive strings byte by byte.
**Learning:** In Go, standard string comparison operators (`==` and `!=`) short-circuit upon finding the first mismatching byte. This creates a measurable timing difference depending on how many leading bytes match, exposing secrets to timing attacks.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` when comparing sensitive strings like passwords, authentication tokens, and credentials. Note that `ConstantTimeCompare` returns 1 if the slices match, and 0 otherwise.
