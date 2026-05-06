## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2026-05-06 - Insecure TLS Certificate Verification fixed

**Vulnerability:** In `internal/netp2p/tls_transport.go` and `internal/netrelay/client.go`, manual TLS certificate verification was missing checks for certificate expiration (`NotBefore` and `NotAfter`), relying solely on certificate fingerprints (`sha256`).

**Learning:** When using `InsecureSkipVerify: true` and custom certificate verification in Go (via `VerifyPeerCertificate` or `VerifyConnection`), expiration dates are not automatically checked by the standard library. Custom verification functions must explicitly assert `time.Now().Before(cert.NotBefore) || time.Now().After(cert.NotAfter)` to ensure expired certificates are rejected. Also, `VerifyConnection` is preferred over `VerifyPeerCertificate` in newer Go versions.

**Prevention:** Ensure all custom TLS certificate verifiers explicitly validate both trust (fingerprint or chain) AND validity periods (expiration dates).
