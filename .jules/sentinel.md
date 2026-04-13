## 2024-05-18 - Fix predictable RNG fallback
**Vulnerability:** Predictable fallback string returned from `randomHex` and `randomKey` when `crypto/rand.Read` entropy generation fails.
**Learning:** Returning a predictable or time-dependent fallback string when cryptography random source fails compromises the entire purpose of generating a secure random string.
**Prevention:** In functions where cryptographic security is critical (like token or seed generation), failure of `crypto/rand` should lead to a panic to fail securely (fail-closed constraint) instead of returning a predictable fallback.
