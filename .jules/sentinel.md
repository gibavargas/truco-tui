## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2026-06-26 - [Custom TLS Certificate Verification]
**Vulnerability:** Missing expiration check and potential timing attacks during manual TLS fingerprint verification.
**Learning:** When `InsecureSkipVerify: true` is used with `VerifyPeerCertificate`, expiration (`NotBefore`/`NotAfter`) isn't implicitly checked. Timing attacks can leak fingerprint characters.
**Prevention:** Always use `VerifyConnection` over `VerifyPeerCertificate` to validate certificate expiration, and use `crypto/subtle.ConstantTimeCompare` for sensitive string comparisons like fingerprints.
## 2026-06-26 - [Cryptographic Entropy Failure Handling]
**Vulnerability:** Silent fallback or error propagation on `crypto/rand` failures during security token and key generation.
**Learning:** If a cryptographic operation fails to read from the entropy source, the application could proceed with predictable or zero-valued secrets if errors are ignored or mishandled upstream.
**Prevention:** Always panic (`panic("entropy source failed")`) immediately when `crypto/rand.Read` returns an error, ensuring a strict fail-closed security posture.
