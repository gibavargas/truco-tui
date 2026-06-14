## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-03-24 - [Fix timing attack vulnerabilities in token comparisons]
**Vulnerability:** Timing attack vulnerability due to using standard string equality operator (`!=`) for checking sensitive token strings (`HostAdminToken` and `joinMsg.Token`).
**Learning:** Standard string comparisons stop at the first differing byte, allowing an attacker to theoretically guess valid tokens byte-by-byte by observing the time taken to process the request.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` (casting strings to `[]byte` and checking `!= 1`) for any sensitive authentication tokens, passwords, or secrets to ensure constant comparison time regardless of input validity.
