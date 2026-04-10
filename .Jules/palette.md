## 2026-04-10 - Accessible Playing Cards
**Learning:** When building complex visual components like playing cards out of multiple HTML text elements (e.g., rank, suit, rank-inverted), screen readers will read all individual elements sequentially, creating confusion. Decorative textual elements within a component should be hidden with `aria-hidden="true"` while the parent wrapper receives `role="img"` and a unified, localized `aria-label`.
**Action:** Apply this pattern to any future compound visual elements built with text or Unicode symbols that represent a single conceptual item.
