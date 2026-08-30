## 2024-04-14 - Predictable Fallbacks in Cryptographic Entropy Sources
**Vulnerability:** Several functions (`newRelayServer` and `randomHex` in `cmd/truco-relay/main.go`, and `randomKey` in `browser-edition/cmd/httpapi/main.go`) used predictable fallback values (hardcoded strings or timestamps) if `crypto/rand` failed to generate entropy.
**Learning:** Falling back to predictable values when entropy generation fails compromises the security of cryptographic operations, session keys, and secrets. It creates a silent failure where the system appears to work but is fundamentally insecure.
**Prevention:** If an entropy source fails during cryptographic operations or secret generation, the application must panic and fail-closed rather than continuing with insecure fallback values.
## 2024-04-14 - CGO Panic and Unpinned Pointers
**Vulnerability:** A `panic: runtime error: cgo argument has Go pointer to unpinned Go pointer` occurred when passing a struct containing Go slice pointers to CGO. Furthermore, passing `&slice[0]` on an empty slice caused a panic.
**Learning:** `cgocheck` detects and panics when unpinned Go pointers are nested inside structs passed to C because the Go garbage collector can move the memory. In addition, getting the address of the first element of an empty slice throws an out of bounds panic.
**Prevention:** Always verify slice length before accessing `&slice[0]`. When passing Go pointers (like slice backing arrays) to CGO, use `runtime.Pinner` to pin them before passing them, allowing C code to safely use the memory.
