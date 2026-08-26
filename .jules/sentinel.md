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
## 2024-08-26 - CGO Pointer Passing Panic and Empty Slice Access
**Vulnerability:** A runtime panic (`index out of range [0] with length 0`) occurred when taking the address of an empty slice's first element (`&slice[0]`) to pass to CGO. Additionally, passing Go pointers to C without pinning violates Go's pointer passing rules, risking unpinned pointers being garbage collected or moved.
**Learning:** You cannot safely access `&slice[0]` if `len(slice) == 0`. Ensure slices passed to CGO have a minimum capacity of 1, even if they are logically empty. Second, any Go pointers nested inside C structs must be pinned using `runtime.Pinner` to prevent crashes (`cgo argument has Go pointer to unpinned Go pointer`).
**Prevention:** Initialize C-bound slices with a minimum capacity (`max(1, len(cards))`). Always pin Go pointers referenced by C structs using `runtime.Pinner.Pin()` before calling into C, and ensure they are unpinned (`defer p.Unpin()`).
