## 2024-05-24 - Cryptographic Fail-Closed Principles

**Vulnerability:** Found multiple instances where the application fell back to predictable or hardcoded values when `crypto/rand.Read` failed to generate random bytes. In `cmd/truco-relay/main.go`, if `rand.Read` failed, the HMAC secret fell back to a hardcoded string `relay-v2-fallback-secret` and `randomHex` fell back to `rnd-fallback`. In `browser-edition/cmd/httpapi/main.go`, `randomKey` fell back to a timestamp. This allows an attacker to predict secrets, session IDs, and forge signed join tickets if the system's randomness source fails.

**Learning:** It's a common anti-pattern to try to handle cryptographic errors by falling back to non-cryptographic pseudo-randomness or hardcoded secrets to prevent immediate application crashes. However, this creates a severe security vulnerability that operates silently, as the application appears to function normally but all generated "secrets" are predictable by an attacker.

**Prevention:** When generating cryptographic secrets (e.g., using `crypto/rand`), the application must panic and fail-closed if the entropy source fails, rather than falling back to a hardcoded or predictable secret. This ensures that a lack of entropy results in a denial of service (which is visible and actionable) rather than a silent security compromise.
