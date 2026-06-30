## 2025-03-11 - Playing Card Accessibility\n**Learning:** When rendering custom CSS-based playing cards that use repeated characters and decorative unicode symbols for the suit, screen readers will read out confusing, nonsensical strings like '4 ♦ ♦ 4 ♦'. To fix this, mark the entire card container with `role="img"` and an `aria-label` containing the readable label (e.g. '4 de Ouros'), while setting `aria-hidden="true"` on all the child elements containing the visual symbols and duplicating caption texts. \n**Action:** Use this `role="img"` with `aria-label` pattern anytime visual text/unicode combinations act purely as decorative graphical elements rather than meaningful flowing text.

## 2025-03-11 - Compact Input Field Accessibility
**Learning:** Text inputs that rely solely on `placeholder` text (without an explicit, visible `<label>`) are difficult or impossible for screen readers to announce correctly. This pattern is common in compact chat interfaces or single-field forms.
**Action:** Always add an `aria-label` attribute (typically matching the placeholder text) to any input field that lacks a `<label>` to ensure proper accessibility for screen reader users.

## 2025-03-11 - Native Form Validation
**Learning:** Relying on JavaScript for simple form validation (like checking for empty fields) introduces complexity and lacks native accessibility cues for invalid states.
**Action:** Default to using native HTML5 validation attributes like `required` on form inputs. It seamlessly blocks form submission, integrates cleanly with existing JavaScript handlers without needing custom logic, and is natively understood by assistive technology.
