## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-05-13 - [TLS Verification Bypass]
**Vulnerability:** Go's custom TLS verification via `VerifyPeerCertificate` bypassed expiration validation when using `InsecureSkipVerify`.
**Learning:** `InsecureSkipVerify: true` skips *all* built-in validation. Implementing a custom fallback mechanism using `VerifyPeerCertificate` historically provided only raw bytes, leading developers to only check fingerprints and ignore expiration times (`NotBefore`/`NotAfter`).
**Prevention:** Use `VerifyConnection` (Go 1.15+) instead of `VerifyPeerCertificate`. `VerifyConnection` receives the `tls.ConnectionState`, allowing direct access to parsed `PeerCertificates` to manually validate both the fingerprint and expiration dates efficiently.
