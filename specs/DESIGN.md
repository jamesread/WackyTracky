# WackyTracky Design Decisions

This document records product and design decisions that affect the frontend and user experience.

---

## Minimalist UI for power users

The WackyTracky frontend is designed as a **minimalist UI for power users**. The interface avoids in-app help clutter, tooltips for every feature, and inline documentation that would slow down users who already know the system. Documentation for syntax, conventions, and advanced behavior lives in **online documentation** (e.g. README, wiki, or dedicated docs site) rather than in the UI.

---

## Decision: No in-app search syntax documentation

**Status:** Will not be implemented.

**Context:** It was suggested to add in-app hints or a "Search help" link to document search syntax (e.g. `#tag`, `@context`, how to combine terms) so users know how to search effectively.

**Decision:** We will **not** add search syntax documentation or search-help UI elements (toolbar hints, "Search help" links, or examples in the search field) inside the application.

**Rationale:**

- The UI is intentionally **minimalist for power users**. In-app hints and help links add visual noise and maintenance burden without benefiting users who have already learned the syntax from external docs.
- Search behavior (including todo.txt-style `#tag`, `@context`, and any backend-specific operators) will be described in **online documentation** (e.g. project README, docs site, or wiki). Power users typically refer to docs once and then rely on keyboard shortcuts and minimal UI.
- Keeping the search bar free of placeholder hints or help links preserves a clean, fast interface and avoids inconsistency when backend search capabilities evolve (docs can be updated in one place).

**Implications:**

- Maintain and keep up to date the **external documentation** for search (syntax, examples, and any backend-specific behavior).
- This is not an open frontend task; it is closed by design.

---

*Other design decisions will be added below as they are made.*
