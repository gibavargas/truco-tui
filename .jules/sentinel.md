## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-05-12 - Incomplete TLS Verification with Fingerprint Pinning
**Vulnerability:** The application pinned the P2P connection's TLS certificate by exclusively checking the SHA-256 fingerprint in a custom `VerifyPeerCertificate` closure, failing to check the certificate's validity dates (`NotBefore`/`NotAfter`).
**Learning:** Pinned certificates, especially dynamically generated ones, still have validity periods. If a private key for a past pinned certificate is compromised, an attacker can reuse the expired certificate if dates are not validated.
**Prevention:** In Go 1.15+, use `VerifyConnection` instead of `VerifyPeerCertificate` when writing custom validation logic with `InsecureSkipVerify: true`. Ensure you validate the dates of the pinned certificate against the current time.
