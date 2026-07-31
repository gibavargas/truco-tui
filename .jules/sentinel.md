## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-03-24 - Missing Certificate Expiration Validation in Custom TLS Config
**Vulnerability:** The application was using `VerifyPeerCertificate` to validate pinned certificate hashes when `InsecureSkipVerify: true` was set, which fails to check the certificate's `NotBefore` and `NotAfter` fields, meaning expired certificates would be accepted as long as the hash matched.
**Learning:** `VerifyPeerCertificate` only provides the raw certificate bytes, bypassing standard expiration checks. When using custom pinning, both the fingerprint AND the validity window must be validated.
**Prevention:** Always use `VerifyConnection` (available in Go 1.15+) instead of `VerifyPeerCertificate` when implementing custom validation, as it provides access to the parsed `x509.Certificate` allowing explicit expiration checks alongside fingerprint validation.
