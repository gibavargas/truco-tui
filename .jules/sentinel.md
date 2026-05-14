## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-05-14 - Missing Expiration Verification on P2P TLS Connections
**Vulnerability:** The custom TLS verification in `internal/netp2p/tls_transport.go` used `VerifyPeerCertificate` to check fingerprint pinning but failed to validate certificate expiration (`NotBefore`/`NotAfter`), potentially allowing connections from compromised or expired certificates.
**Learning:** When using `InsecureSkipVerify: true` for custom TLS logic (like fingerprint pinning), using `VerifyPeerCertificate` bypasses all standard checks, including time-based validity.
**Prevention:** In Go 1.15+, use `VerifyConnection` (which passes `tls.ConnectionState`) instead of `VerifyPeerCertificate` to explicitly validate `NotBefore` and `NotAfter` bounds against `time.Now()` alongside any custom fingerprint pinning.
