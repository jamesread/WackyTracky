# Main task list UI: minimalism

This spec records design decisions that keep the main task list free of common per-row controls and as minimalist as possible.

---

## Principle: minimal task list

The **main task list** (the list of tasks in list view and search results) must stay **free of common buttons and as minimalist as possible**. Row content is limited to: structural indicators (collapse/expand, next-action, waiting), task content and metadata (tags, contexts, due), and inline edit when active. Actions are reached via **keyboard shortcuts** and **context menu (right-click)** rather than per-row buttons or icons. This keeps the list dense, scannable, and fast for power users who rely on keys and gestures.

---

## Decision: No one-click complete control in the list row

**Status:** Will not be implemented.

**Context:** It is often suggested to add a one-click "complete" control in each task row—for example a checkbox or checkmark icon at the start of the row—so users can mark tasks done without opening a menu or using the keyboard.

**Decision:** We will **not** add a checkbox, checkmark icon, or any other per-row "complete" control to the main task list.

**Rationale:**

- The main task list is intentionally **minimal**. A complete control on every row adds visual and interaction clutter and pushes the list toward a "button bar per task" pattern we explicitly avoid.
- Completion is already supported by **keyboard** (e.g. `d` `d` or Delete when a row is focused) and by **right-click → task details → Mark task done**. Power users can complete tasks quickly without per-row controls.
- Keeping the list free of per-row action buttons preserves a clean, consistent look and leaves room for content and metadata.

**Implications:**

- Completion remains keyboard- and context-menu–driven. Document these flows in the keyboard shortcuts reference and any user-facing docs.
- Do not treat "add a complete checkbox/icon to each row" as an open task; it is closed by design.

---

## Decision: No explicit "open details" control in the list row

**Status:** Will not be implemented.

**Context:** It is often suggested to add an explicit way to open task details from the list row—for example a "…" or info icon on each row, or making single click on the task text open details—so that details are discoverable without right-click.

**Decision:** We will **not** add a per-row "open details" control (e.g. "…" or info icon) to the main task list. Task details remain available via **right-click** and via **keyboard** (e.g. Enter or a dedicated key when a row is focused, as implemented).

**Rationale:**

- The main task list is intentionally **minimal**. An icon or button on every row to open details would contradict the goal of a button-free, minimalist list.
- Details are already reachable via **right-click** and **keyboard**. Power users learn these once; the shortcuts reference and docs can describe "Right-click for details" and the relevant keys.
- Avoiding per-row action icons keeps the list scannable and consistent with the decision to avoid a per-row complete control.

**Implications:**

- Task details stay discoverable through context menu and keyboard; document these in the shortcuts reference and user-facing documentation.
- Do not treat "add a … or info icon to each row to open details" as an open task; it is closed by design.

---

*Further decisions about the main task list UI may be added below.*
