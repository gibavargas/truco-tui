## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2026-05-08 - Use VerifyConnection instead of VerifyPeerCertificate for custom TLS validation
**Vulnerability:** When using `InsecureSkipVerify: true` for custom TLS logic like certificate fingerprint pinning, `VerifyPeerCertificate` only provides the raw bytes of the certificates. Failing to manually decode and check the expiration dates (`NotBefore`/`NotAfter`) can allow expired certificates to be accepted.
**Learning:** `VerifyPeerCertificate` is harder to use securely because the developer must manually parse the ASN.1 DER data to get the standard `x509.Certificate` fields.
**Prevention:** In Go 1.15+, use `VerifyConnection(cs tls.ConnectionState) error` instead. It provides fully parsed `cs.PeerCertificates` where standard fields like `NotBefore` and `NotAfter` are immediately accessible and can be validated efficiently alongside custom checks.
