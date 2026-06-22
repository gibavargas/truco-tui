## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-05-18 - Timing Attack in Credential Verification
**Vulnerability:** Found `mem.Credential != req.PeerCredential` and `mem.Credential != h.Credential` using standard string equality in `cmd/truco-relay/main.go`. This makes it susceptible to timing attacks, as string comparisons fail fast.
**Learning:** Checking credentials using `==` or `!=` leaks information about string matches byte-by-byte. For security tokens/credentials, this allows attackers to guess values over time.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1` when comparing sensitive tokens, credentials, passwords, or hashes.
