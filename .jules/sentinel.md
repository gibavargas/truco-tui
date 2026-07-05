## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2023-10-18 - TLS Pinned Certificate Expiration Bypass
**Vulnerability:** Pinned TLS connections ignored certificate expiration dates because they used `VerifyPeerCertificate` to only check the SHA-256 fingerprint while bypassing standard verification with `InsecureSkipVerify: true`.
**Learning:** Raw fingerprint pinning validates the identity (tampering with dates alters the hash), but it fails to defend against replay or reuse of a legitimately expired certificate if expiration isn't checked manually.
**Prevention:** Always use `VerifyConnection` (Go 1.15+) instead of `VerifyPeerCertificate` when using `InsecureSkipVerify: true` to explicitly validate `cert.NotBefore` and `cert.NotAfter` alongside the fingerprint.
