## 2024-04-06 - [Decorative Unicode Accessibility]
**Learning:** Decorative Unicode characters (like ▶, 💬, ↻, ⎋, ⟳) used in button text or labels can be read out loud by screen readers, causing confusion.
**Action:** Wrap these characters in `<span aria-hidden="true">` to hide them from screen readers.
