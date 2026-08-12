## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-08-12 - Sensitive Tokens Stored as Plaintext Map Keys
**Vulnerability:** Sensitive replacement invite tokens were stored in plaintext as keys in the `replaceInvites` Go map, which risks memory leakage and timing attacks.
**Learning:** Storing sensitive plaintext tokens directly in memory structures like hash maps introduces a risk of map-lookup timing attacks and exposes secrets in memory dumps.
**Prevention:** Always hash sensitive tokens using strong cryptographic hashes like `crypto/sha256` before using them as keys in memory maps to provide a defense-in-depth security layer.
