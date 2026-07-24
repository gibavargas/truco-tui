## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-07-24 - Timing Attacks in Token Verification
**Vulnerability:** Several authentication paths (`cmd/truco-relay/main.go` and `internal/netp2p/host.go`) compared sensitive strings (`HostAdminToken`, `Credential`, `Token`) using standard equality (`!=`). This exposed the application to timing attacks, allowing potential inference of valid token characters based on the string comparison runtime.
**Learning:** Standard string equality (`==` or `!=`) short-circuits on the first mismatched character. This is unsafe for authenticating sensitive strings against untrusted inputs.
**Prevention:** Always cast sensitive strings to `[]byte` and use `crypto/subtle.ConstantTimeCompare(a, b) == 1` when evaluating equality in authentication or cryptographic pathways. Note that Python's `ConstantTimeCompare` checks for `== 0`, but Go's implementation returns `1` if the two slices are equal and `0` otherwise, so the logic must check `== 1` for equality (or `== 0` for inequality, matching `!=`).
