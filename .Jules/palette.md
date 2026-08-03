## 2025-03-11 - Playing Card Accessibility\n**Learning:** When rendering custom CSS-based playing cards that use repeated characters and decorative unicode symbols for the suit, screen readers will read out confusing, nonsensical strings like '4 ♦ ♦ 4 ♦'. To fix this, mark the entire card container with `role="img"` and an `aria-label` containing the readable label (e.g. '4 de Ouros'), while setting `aria-hidden="true"` on all the child elements containing the visual symbols and duplicating caption texts. \n**Action:** Use this `role="img"` with `aria-label` pattern anytime visual text/unicode combinations act purely as decorative graphical elements rather than meaningful flowing text.

## 2023-10-27 - Implicit Labels vs ARIA Labels
**Learning:** This application uses implicit labels for most forms (e.g. `<label><span>Text</span><input></label>`). This is WCAG compliant. However, compact UI areas like the chat bar completely omit visible `<label>` text, relying solely on placeholders. Placeholders are not a reliable substitute for screen readers.
**Action:** When a visible label must be omitted for design reasons (like inline chat inputs), always add an `aria-label` attribute directly to the `<input>`, ideally reusing the localized placeholder string if it's descriptive enough.

## 2023-10-27 - TestPlayFaceDownMasksBrowserSnapshot panic
**Learning:** The test `TestPlayFaceDownMasksBrowserSnapshot` frequently panics locally due to a CGO logic binding error in the AI Engine when simulating CPU turns during test advancement. This does not block or invalidate frontend accessibility fixes and should be ignored for UI-only updates.
**Action:** Ignore this panic when verifying frontend-only modifications to the browser app.

## 2023-10-27 - CGO Unpinned Pointer Panic
**Learning:** Passing Go slice pointers directly to C functions (e.g. `&slice[0]`) panics with `cgo argument has Go pointer to unpinned Go pointer` in Go 1.21+ if the Go garbage collector can move the underlying array.
**Action:** When passing dynamically sized arrays to C, use `C.malloc` to allocate C memory, copy the data, pass the C pointer, and ensure it is freed with `C.free` using `defer`.
