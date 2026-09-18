## 2026-09-18 - Adding aria-label to compact text inputs
**Learning:** Compact text inputs that rely solely on placeholders for visual context are not accessible to screen readers without an explicit label or aria-label.
**Action:** Always ensure that text inputs without an explicit `<label>` tag include an `aria-label` attribute, especially for ubiquitous compact components like chat input fields.
