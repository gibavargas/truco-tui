## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-05-03 - Missing Expiration Check in InsecureSkipVerify
**Vulnerability:** The application used `InsecureSkipVerify: true` to bypass default verification and do ad-hoc fingerprint pinning, but used `VerifyPeerCertificate` instead of `VerifyConnection`. This meant that standard certificate expiration fields (`NotBefore` and `NotAfter`) were completely ignored.
**Learning:** Using custom certificate validation with `VerifyPeerCertificate` without explicitly checking standard properties like expiration date ignores default protections. `VerifyConnection` is preferred since it gives access to parsed certs to perform full validation when `InsecureSkipVerify` disables it.
**Prevention:** When using `InsecureSkipVerify: true` for custom TLS pinning (like ad-hoc P2P), use `VerifyConnection` (Go 1.15+) instead of `VerifyPeerCertificate`, and explicitly validate certificate expiration (`NotBefore`/`NotAfter`) along with the custom fingerprint or criteria.
