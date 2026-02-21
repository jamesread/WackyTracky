# Vue-based confirmation and input dialogs

This spec defines how confirmation and input flows should be implemented so that browser dialogs are not used and the UI stays consistent. Following it reduces the risk of blocking or inconsistent experiences (e.g. using `window.prompt` or `window.confirm`).

---

## Principle: Vue-based dialogs only

**Confirmation and input dialogs** must be implemented as **Vue-based UI**, not with browser APIs. Do **not** use `window.confirm`, `window.prompt`, or `window.alert` for any user-facing confirmation or text input. Use in-app overlays and dialogs built with Vue (templates, refs, and existing patterns or components) so that behavior, styling, and accessibility stay under application control and consistent with the rest of the app.

---

## Reuse existing components and patterns

- **Before adding new dialog UI**, check for existing components or patterns in the project that can be reused or adapted.
- **Input dialogs** (e.g. “Enter a name”): Prefer reusing or extending the pattern used by **CreateListDialog** (overlay + inner dialog, `v-model` for open state, form with label + input + Cancel/Submit, inline error message, `role="dialog"`, `aria-modal="true"`, `aria-labelledby`). Build a similar Vue component or inline overlay for new flows (e.g. “Save saved search” name) instead of using `window.prompt`.
- **Confirmation dialogs** (e.g. “Delete X?”): Reuse the pattern used in the app for confirmations (e.g. overlay + dialog with title, message, Cancel + primary action, controlled by a ref and `v-if`). Build a small Vue component or an inline overlay in the relevant view; do not use `window.confirm`.
- **Do not introduce new npm dependencies** for dialogs, modals, or confirmations unless the user explicitly requests or adds them. Use only Vue and existing project dependencies.

---

## Implementation checklist

When you need a confirmation or a single-field input from the user:

1. **Avoid browser APIs:** Do not call `window.confirm`, `window.prompt`, or `window.alert`.
2. **Reuse first:** Look for an existing Vue dialog component or an existing overlay/dialog pattern in the codebase (e.g. CreateListDialog, confirm overlays in ListView, list-options or repo-status dialogs in App.vue). Reuse or copy that pattern.
3. **Implement in Vue:** Add either:
   - A **reusable component** (e.g. a generic confirm dialog or input dialog that takes props and emits result), or
   - An **inline overlay** in the same view (a `v-if`-controlled overlay + inner dialog with title, content, and actions), following the same structure as existing dialogs (overlay with `@click.self` to close, inner div with `role="dialog"` and `aria-labelledby`, buttons for Cancel and primary action).
4. **Accessibility:** Use `role="dialog"`, `aria-modal="true"`, and `aria-labelledby` pointing to the dialog title. Ensure the primary action and Cancel are focusable and that Escape (or an explicit Cancel) closes the dialog when appropriate.
5. **No new dialog libraries:** Implement with Vue only; do not add npm packages for modals or confirmations unless the user explicitly adds them.

---

## Rationale

- **Consistency:** In-app Vue dialogs share styling, behavior, and placement with the rest of the app; browser dialogs look and behave differently and block the main thread.
- **Control:** Vue dialogs can show validation, existing options, or retry state; `window.prompt` and `window.confirm` cannot.
- **No extra dependencies:** Reusing existing patterns keeps the bundle small and avoids adding dialog libraries unless the user explicitly wants them.

---

## Scope

- Applies to all **confirmation** flows (e.g. “Delete saved search?”, “Mark task done?”) and **single-field input** flows (e.g. “Name for this saved search”) that would otherwise be implemented with `window.confirm`, `window.prompt`, or `window.alert`.
- Does not require a single shared “dialog library”; reuse can be by copying the structure of CreateListDialog or existing confirm overlays and adapting props/slots as needed.
- “User explicitly adds them” means the user has explicitly requested or added an npm dependency; normal feature work should not add dialog/modal packages on its own.

---

*This spec should be treated as a normative guideline for frontend implementation and code review.*
