## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.

## 2024-05-18 - Timing Attack Vulnerability in Token Comparisons
**Vulnerability:** String comparison operators (`!=` or `==`) were being used to compare authentication tokens (`req.HostAdminToken != sess.AdminToken` and `joinMsg.Token != h.token`).
**Learning:** Standard string comparisons stop at the first differing character, revealing information about the matching prefix through execution time, which can be exploited in a timing attack to guess tokens character by character.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare([]byte(given), []byte(expected)) == 1` when checking passwords, tokens, API keys, MACs, or any other sensitive secrets.
