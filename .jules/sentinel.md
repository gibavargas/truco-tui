## 2025-01-21 - [CRITICAL] Fail-closed behavior for cryptographic operations

**Vulnerability:**
The codebase contained predictable fallbacks for random number generation when `crypto/rand.Read()` failed. Specifically, the relay server (`cmd/truco-relay/main.go`) fell back to hardcoded strings like `"relay-v2-fallback-secret"` and `"rnd-fallback"`, while the browser edition API (`browser-edition/cmd/httpapi/main.go`) used the current timestamp (`time.Now().UnixNano()`). If the system's entropy pool became depleted, these fallbacks would result in deterministic keys, session IDs, and tokens, completely compromising the application's security.

**Learning:**
Go's `crypto/rand.Read` is generally reliable, but system-level entropy depletion can still cause failures (especially in containerized or embedded environments). Using predictable fallbacks in these scenarios is extremely dangerous. It is a critical security pattern in Go (and other languages) to **fail-closed** (e.g., panic) when secure random number generation fails, preventing the application from running in a compromised state. This ensures we do not unknowingly operate with predictable secrets.

**Prevention:**
Always verify the error returned by `crypto/rand.Read`. If it fails, the application **must** panic or immediately terminate with a fatal error instead of providing a predictable or weak fallback value.
