
## 2024-04-13 - Add required attributes for native validation
**Learning:** For web forms (like setup screens and chat inputs), native HTML5 validation constraints (`required` attribute) provide immediate, accessible feedback to screen readers and keyboard users without relying on entirely custom JavaScript logic. The browser handles focusing the empty field and reading out "Please fill out this field", which creates a more predictable experience.
**Action:** When creating text inputs that must not be empty (e.g. player name, chat message), add the `required` attribute. Avoid adding it to dynamically handled setup fields (like relay-url or invite-key) where it might break specific connection flows.
