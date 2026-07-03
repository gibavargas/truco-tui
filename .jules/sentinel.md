## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-05-28 - Insecure TLS Configuration in netp2p
**Vulnerability:** The `tlsClientConfig` in `internal/netp2p/tls_transport.go` used `VerifyPeerCertificate` alongside `InsecureSkipVerify: true` to perform custom certificate fingerprint validation.
**Learning:** When `InsecureSkipVerify` is true, Go does not check the certificate's expiration date (`NotBefore`/`NotAfter`) by default. Using `VerifyConnection` is preferred over `VerifyPeerCertificate` in modern Go as it allows explicitly validating both the fingerprint and the expiration of the certificate, closing a security gap where expired self-signed certificates would still be accepted.
**Prevention:** Always use `VerifyConnection` instead of `VerifyPeerCertificate` when implementing custom certificate validation with `InsecureSkipVerify: true` to explicitly handle standard validations like expiration.
