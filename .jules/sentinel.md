## 2024-04-16 - [CRITICAL] Fix insecure crypto entropy fallback

**Vulnerability:** Go applications (like `httpapi` and `truco-relay`) had an insecure fallback pattern where if `crypto/rand` failed to read entropy, they silently fell back to predictable states (e.g., timestamps or hardcoded strings) for generating critical IDs or secrets.
**Learning:** `crypto/rand` is essential for secure token and secret generation. Falling back to non-cryptographically secure sources allows attacks to guess the generated values, leading to session hijacking or access compromise.
**Prevention:** Always fail securely (fail-closed) by calling `panic` when cryptographic entropy sources like `crypto/rand` fail to provide the requested bytes. Never use timestamps or hardcoded values as fallback entropy.
