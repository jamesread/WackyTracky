# Toast feedback for async operations

This spec defines a consistent pattern for user feedback so that success and failure of async operations (especially API calls) are never silent. Following it reduces the risk of bugs where users assume an action succeeded or are left unaware that it failed.

---

## Principle: use toasts for success and failure

**Toast notifications** must be used to report both **success** and **failure** for any user-triggered operation that performs async work (e.g. API calls). The app provides a non-blocking toast (e.g. "Task added", "Could not save: [reason]") that appears briefly so the user gets immediate, unambiguous feedback without modal dialogs or inline error UI in every screen.

- **Success:** After an operation completes successfully, show a short success toast (e.g. "Task added", "List renamed", "Saved") unless the UI already makes the outcome obvious (e.g. a dialog closing and the list updating).
- **Failure:** On API or client error, show an error toast that includes a brief reason (e.g. "Could not save: [reason]"). Keep the user in the same context (e.g. keep the dialog open, keep draft content) so they can correct and retry. Use a consistent pattern for the reason: `e?.message || String(e)` or an equivalent so the user sees something actionable.

Operations in scope include but are not limited to: create/update/complete task, rename/delete list, save task metadata (notes, priority, wait until, due date), load lists, search, and any other action that can fail due to network or validation.

---

## Implementation checklist

When adding or changing an async, user-triggered operation:

1. **Success path:** Call the app’s toast helper (e.g. `showToast(message)`) with a short success message after the operation succeeds, unless the result is already obvious from the UI (e.g. navigation or a clear state change).
2. **Failure path:** In every `catch` (or equivalent) that handles the operation, call the toast helper with an error message (e.g. `showToast('Could not …: ' + reason, 'error')`). Do not swallow errors with an empty `catch` or a comment like "keep dialog open" without also showing a toast.
3. **Context on failure:** Prefer keeping the user in the same context on failure (e.g. leave the dialog open, retain form values) so they can fix and retry. The toast tells them *that* it failed; the UI state lets them *retry* without re-entering data.

---

## Rationale

- **Reliability and trust:** GTD and task apps depend on users trusting that actions were applied. Silent failures (e.g. save error with no message) lead to wrong assumptions and duplicate or lost work. Success toasts confirm that the system accepted the action.
- **Consistency:** Using toasts for both success and failure in one place makes it easier to apply the same pattern everywhere and to review code for missing feedback.
- **Low friction:** Toasts are non-blocking and brief, so they don’t slow down power users while still informing them.

---

## Scope and exceptions

- **Loading/initial fetch:** If a *load* fails (e.g. "could not load lists"), show an error toast and/or an inline error state (e.g. "Could not load lists. Please try again.") so the user knows the view is in an error state, not simply empty.
- **Background or non–user-triggered work:** For operations not directly triggered by the user (e.g. a background sync), use the same pattern if the result should be visible; otherwise document that feedback is deferred or omitted by design.
- **Obvious outcomes:** When success is already obvious (e.g. dialog closes and the list visibly updates), a success toast is optional but still recommended for consistency and accessibility.

---

*This spec should be treated as a normative guideline for frontend implementation and code review.*
