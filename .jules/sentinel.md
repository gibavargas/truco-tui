## 2025-04-11 - [Predictable Fallback Secrets on Entropy Failure]
**Vulnerability:** Found multiple instances where `crypto/rand.Read` failures resulted in falling back to predictable values like hardcoded strings (`"relay-v2-fallback-secret"`, `"rnd-fallback"`) or predictable timestamps (`time.Now().UnixNano()`).
**Learning:** In Go, fallback logic for `crypto/rand` is almost always a security flaw. If the entropy source fails, the system is fundamentally compromised and generating predictable "random" values silently introduces severe vulnerabilities (predictable session IDs, relay secrets, etc).
**Prevention:** Always fail-closed (e.g., `panic` or fatal error) if `crypto/rand` fails. Never fall back to weak entropy or hardcoded strings.
