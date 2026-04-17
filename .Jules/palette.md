
## 2026-04-17 - Accessible CSS Playing Cards
**Learning:** Decorative CSS-based components that use Unicode characters and text to form a visual graphic (like playing cards) are read out confusingly by screen readers (e.g., reading rank and suit multiple times or reading raw symbols).
**Action:** When rendering such components, apply `role="img"` and a descriptive `aria-label` to the main container, and set `aria-hidden="true"` on all child elements. This consolidates the element into a single, understandable image for screen readers.
