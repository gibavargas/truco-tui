## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-05-15 - [HIGH] Fix P2P pinning expiration verification bypass
**Vulnerability:** The P2P self-signed TLS implementation used `VerifyPeerCertificate` to validate certificate fingerprints but didn't check expiration dates (`NotBefore`/`NotAfter`).
**Learning:** Raw certificate fingerprint pinning prevents expiration tampering, but expiration should still be enforced explicitly to ensure stale pinned certificates are rejected. Also, fingerprint checking should be time-constant to prevent timing attacks.
**Prevention:** Use `VerifyConnection` over `VerifyPeerCertificate` to natively check the certificate's `NotBefore` and `NotAfter` limits against `time.Now()` alongside securely doing `ConstantTimeCompare` on the fingerprint.
