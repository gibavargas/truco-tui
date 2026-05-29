## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2025-02-14 - Fix Incomplete TLS Validation
**Vulnerability:** Missing TLS certificate expiration check when using InsecureSkipVerify: true in P2P transport.
**Learning:** Using InsecureSkipVerify with VerifyPeerCertificate allows pinning fingerprints but fails to natively check NotBefore/NotAfter times, potentially allowing expired certificates to be accepted. VerifyConnection provides the full ConnectionState and is the recommended way in Go 1.15+ to handle custom certificate verification.
**Prevention:** Always use VerifyConnection instead of VerifyPeerCertificate when custom validation is required with InsecureSkipVerify: true, and explicitly check cs.PeerCertificates[0].NotBefore and NotAfter.
