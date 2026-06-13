## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## $(date +%Y-%m-%d) - Fix timing attack vulnerabilities in token comparison

**Vulnerability:** Secure authentication tokens were compared using standard string equality operators (`!=`), which can be vulnerable to timing attacks. This could allow an attacker to guess the token byte by byte by measuring the time taken for the comparison to return.
**Learning:** Even internal security mechanisms like host recovery tokens and relay admin tokens need to be compared securely. The string-based check `joinMsg.Token != h.token` or `req.HostAdminToken != sess.AdminToken` was used out of convenience but represents a potential timing attack vector.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(a), []byte(b)) != 1` when comparing sensitive tokens, secrets, passwords, or API keys.
