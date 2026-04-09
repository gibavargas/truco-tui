## 2024-06-25 - Prevent Screen Reader Reading out Decorative Icons
**Learning:** Decorative Unicode characters (like ▶, 💬, ↻, ⎋) used in button text or labels cause screen readers to read them out loud, creating a confusing and unpleasant user experience for visually impaired users.
**Action:** Always wrap these decorative Unicode characters in `<span aria-hidden="true">` to prevent screen readers from reading them out loud.
