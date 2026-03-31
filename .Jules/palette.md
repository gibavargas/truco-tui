## 2024-05-24 - Hiding inline decorative Unicode characters
**Learning:** Decorative Unicode characters (like `▶`, `💬`, `⟳`, `⎋`) used directly in button text or labels are often read out loud by screen readers, creating a confusing or noisy experience.
**Action:** Always wrap these visual-only characters in `<span aria-hidden="true">` to ensure screen readers announce the action cleanly.