## 2026-03-30 - Missing aria-labels on chat inputs
**Learning:** Chat inputs across the application (in both the game and lobby views) were lacking associated `<label>` elements or `aria-label` attributes. Since these inputs rely on placeholders for context, screen readers couldn't properly identify them without aria-labels.
**Action:** Always ensure that inputs without a visible `<label>` element have a descriptive `aria-label` attribute (often matching the placeholder text) so screen reader users understand what the input is for.
