## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2023-10-27 - Insecure TLS Config
**Vulnerability:** The application used `VerifyPeerCertificate` to validate certificate fingerprints but didn't check expiration dates (`NotBefore`/`NotAfter`).
**Learning:** `VerifyPeerCertificate` doesn't automatically validate certificate dates.
**Prevention:** Use `VerifyConnection` instead when custom TLS validation logic is needed (like pinning). It allows checking expiration alongside custom criteria to prevent accepting stale certificates.
