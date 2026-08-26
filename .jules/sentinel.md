## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-08-26 - Timing Attacks on String Comparisons of Secrets
**Vulnerability:** Authentication tokens and credentials were compared using standard string equality operators (`==` and `!=`) across relay logic and network peer connections.
**Learning:** Standard string equality comparisons return early upon finding the first non-matching byte. This exposes the application to timing attacks where an attacker can incrementally guess the token by measuring the time it takes for the application to reject it.
**Prevention:** Always use `crypto/subtle.ConstantTimeCompare` when comparing secrets, passwords, or tokens in Go to ensure constant time execution regardless of the input data.
## 2024-08-26 - CI Linker Errors from CGO Dependencies
**Vulnerability:** CI pipelines failed to link Go binaries that depended on a C++ shared library (`truco_ai`) because the compilation step for the C++ library was missing or ordered incorrectly.
**Learning:** When dealing with CGO dependencies in CI pipelines, ensure that all C/C++ libraries are built *before* any Go toolchain steps (like `go vet`, `go test`, or `go build`) are executed, otherwise the linker will fail.
**Prevention:** Always check `ci.yml` to ensure native compilation steps explicitly precede Go checks.
