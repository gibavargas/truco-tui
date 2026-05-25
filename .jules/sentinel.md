## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2026-05-25 - P2P TLS Validation and Token Timing
**Vulnerability:** The P2P TLS transport bypassed certificate validity period checking because `VerifyPeerCertificate` only provides the raw bytes, and it was used with `InsecureSkipVerify: true`. Additionally, the host join token was verified using standard string comparison `!=`, which is vulnerable to timing attacks.
**Learning:** When using custom self-signed certificates with `InsecureSkipVerify: true`, `VerifyConnection` must be used instead of `VerifyPeerCertificate` to properly validate the `NotBefore` and `NotAfter` fields of the parsed certificate. Additionally, any authentication tokens must be compared using `crypto/subtle.ConstantTimeCompare`.
**Prevention:** Always ensure custom TLS validation checks the certificate expiration. Use constant-time comparison for all security-sensitive tokens.
