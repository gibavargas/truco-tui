## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2026-05-17 - Fix custom TLS certificate verification missing expiration checks
**Vulnerability:** The custom TLS verification logic for self-signed certificates used `InsecureSkipVerify: true` and `VerifyPeerCertificate` to check the certificate's SHA-256 fingerprint but completely ignored its validity period (`NotBefore` and `NotAfter`).
**Learning:** In Go, when `InsecureSkipVerify: true` is set, `VerifyPeerCertificate` does not receive parsed certificates in `rawCerts`, making it tedious to parse and check expiration manually. Using `VerifyConnection` (available in Go 1.15+) provides access to the parsed certificates via `ConnectionState.PeerCertificates`, making it trivial to properly enforce expiration.
**Prevention:** Always use `VerifyConnection` instead of `VerifyPeerCertificate` when custom verification logic is required alongside `InsecureSkipVerify: true`, and explicitly validate the `NotBefore` and `NotAfter` fields to ensure expired certificates are rejected.
