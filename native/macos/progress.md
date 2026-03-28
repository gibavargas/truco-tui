# Native macOS Progress

Status: the macOS client has been moved off the mock snapshot bridge and onto the real C ABI path. Browser Edition and Wails are now the main cross-platform targets. This file is the handoff for the remaining macOS-only work.

## Done

- `TrucoCoreBridge.swift` now calls the real `TrucoCore*` exported symbols.
- `Models.swift` was aligned to the `SnapshotBundle` shape.
- `GameView.swift`, `CardView.swift`, and `TrucoMacApp.swift` were updated to use the new runtime-backed state.

## Still Needed on macOS

- Build the app on macOS and fix any Swift / linker / ABI mismatches.
- Verify the generated C ABI symbols still match the Swift declarations.
- Run offline gameplay end-to-end on macOS.
- Run online host/join end-to-end on macOS.
- Verify chat, host vote, replacement invite, locale switching, new hand, and reset flows.
- Check that snapshot decoding matches the live runtime output under real gameplay.
- Confirm the UI state updates correctly when the runtime emits events and errors.

## Suggested Mac Verification

```bash
go test ./internal/appcore
```

Then on macOS:

```bash
xcodebuild -project native/macos/TrucoMac.xcodeproj -scheme TrucoMac -configuration Debug build
```

If the app still uses a generated FFI or bridge target, verify the native library / module build step before opening the app.

## Notes

- Do not touch WinUI or GTK from this handoff.
- Treat Browser Edition and Wails as the interoperability reference; macOS should match their supported feature set, not diverge from it.
- If a mismatch appears, prefer fixing the shared runtime contract first, then adjust the Swift UI to match it.
