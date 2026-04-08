## 2024-04-08 - Decorative Unicode Characters in UI Controls
**Learning:** Found multiple instances where decorative unicode characters (like ▶, 💬, ⟳, ⎋) were placed inside interactive UI elements (buttons) without being explicitly hidden from screen readers. This can cause screen readers to announce the characters in unpredictable or confusing ways, degrading the experience.
**Action:** Always wrap decorative unicode characters in buttons or labels with `<span aria-hidden="true">` to ensure screen readers only announce the meaningful, localized text label.
