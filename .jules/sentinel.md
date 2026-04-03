## 2025-04-03 - Insecure Randomness Fallbacks
**Vulnerability:** Core services (relay and HTTP API) fell back to insecure, predictable random values (hardcoded strings or timestamps) when the CSPRNG (`crypto/rand`) failed to generate entropy.
**Learning:** This "silent fallback" pattern defeats the purpose of cryptographic randomness and creates a false sense of security, potentially allowing token prediction or session hijacking if entropy is exhausted.
**Prevention:** Always fail securely by panicking or returning an error when the system cannot generate secure randomness. Never fall back to predictable methods for security-sensitive tokens.
