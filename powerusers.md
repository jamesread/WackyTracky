# Power user suggestions

This document lists suggestions to make WackyTracky more appealing to power users and highly technical people who are often frustrated by other task-tracking tools (locked-in data, slow UIs, no keyboard control, no API, or no way to script or automate).

The app already targets this audience with vim-style bindings, minimal UI, todo.txt plain-text storage, git-friendly data, and a real API. The items below are enhancements that would deepen that appeal.

---

## Already strong for power users

- **Keyboard-first**: vim-style j/k navigation, `dd` to complete, `yy`/`p` to yank/paste, `?` for shortcuts, Ctrl+K/F for search, minimal point-and-click.
- **Minimal list UI**: No per-row checkboxes or action icons; actions via keyboard and right-click (see [main-task-list-ui](../specs/main-task-list-ui.md)).
- **Plain-text backend**: todo.txt format; data is editable in any editor, syncable with Git/Dropbox/Syncthing.
- **Git integration**: Repo status for the todotxt directory is available from the UI (Options → Repo status).
- **Saved searches**: Named saved searches with todo.txt-style query (tags, contexts, exclude with `-`).
- **Hide-at-times**: Tag/context visibility rules using [expr](https://github.com/expr-lang/expr) (e.g. hide `@work` on weekends).
- **ConnectRPC API**: Full gRPC/Connect API for lists, tasks, search, metadata, TPPs, rule testing, version, repo status.
- **Self-hosted, no tiers**: One version, no “pro” paywall, no telemetry.

---

## Suggestions (prioritized by impact for power users)

### 1. Document the API for scripters

**Problem:** Power users want to add tasks from the shell, cron, or scripts. The API exists (Connect/gRPC) but is not documented for HTTP callers.

**Suggestion:**

- Publish API docs (OpenAPI/Connect schema or a “Scripting” doc) that show:
  - Base URL and auth (if any).
  - How to call RPCs over HTTP (Connect protocol supports HTTP/JSON).
  - Example `curl` or `fetch` for: add task, list tasks, search, mark done.
- Optionally provide a small example script (e.g. “add task from stdin”) in the repo or docs.

**Impact:** Enables automation (CI, cron, shell aliases, other tools) without reverse-engineering the frontend.

---

### 2. Official or documented CLI

**Problem:** Many power users live in the terminal. They want to quick-add a task or run a search without opening the browser.

**Suggestion:**

- Provide a small CLI (e.g. `wt` or `wackytracky`) that:
  - Accepts server URL (and optionally auth) via env or config.
  - Supports: add task, list/search (with same query syntax as UI), mark done, optionally list lists.
- Implement it in Go (same repo) or document a thin wrapper (e.g. shell script calling the HTTP API) so users can add tasks from the command line or from scripts.

**Impact:** Terminal-native workflow; integrates with existing shell habits and automation.

---

### 3. Bulk operations (multi-select)

**Problem:** Power users often need to complete, move, or retag many tasks at once. Doing it one-by-one is tedious.

**Suggestion:**

- Add multi-select in the list view:
  - Shift+click or Ctrl/Cmd+click to extend/range-select; optional “select all on screen”.
  - Toolbar or context menu: “Mark selected done”, “Move selected to list”, “Set due date / priority / tag / context” for selection.
- Backend: either existing RPCs in a loop or new batch RPCs (e.g. `DoneTasks(ids)`, `MoveTasks(ids, list_id)`) for efficiency and atomicity.

**Impact:** Makes cleanup and reorganization fast for users with many tasks.

---

### 4. Search syntax documentation (external)

**Problem:** Power users rely on search; they need to know exact syntax (e.g. `#tag`, `@context`, `-term` to exclude) and any backend-specific behavior.

**Suggestion:**

- Keep syntax out of the in-app UI (per [DESIGN](../specs/DESIGN.md)) but ensure the **docs site** has a dedicated “Search” page that covers:
  - Terms, tags, contexts, exclude, and how they combine.
  - Any limits (e.g. freetext vs structured) and which backends support what.
- Link to this from the main docs index and from the README so scripters and power users can find it once and rely on it.

**Impact:** Reduces trial-and-error and builds trust that the app is built for people who care about precise behavior.

---

### 5. Customizable keyboard shortcuts

**Problem:** Some power users prefer different bindings (e.g. non-vim) or need to avoid conflicts with browser/OS shortcuts.

**Suggestion:**

- Add a “Keyboard” section in Options (or a config file for self-hosted) where users can:
  - Remap critical actions (e.g. complete, edit, search, move up/down) to different keys.
  - Optionally choose a preset: “Vim” (current) vs “Default” (e.g. arrow keys, Enter to edit, Delete to complete).
- Store preferences in localStorage or backend so they persist across devices if the user is logged in.

**Impact:** Respects different muscle memory and environments (e.g. non-vim users, non-QWERTY).

---

### 6. Export and backup

**Problem:** Power users want to backup data, migrate, or run external analytics without scraping the UI.

**Suggestion:**

- **Export:** An “Export” action (or API) that returns:
  - Current list/task set (e.g. JSON or todo.txt format) for the current list or for search results.
  - Optional: full dump of all lists and tasks (and done) for backup.
- **Backup:** Document that with the todo.txt backend, the data directory itself is the backup (and how to copy or archive it); for other backends, document or provide export so users can snapshot their data.

**Impact:** Trust and control; no fear of lock-in.

---

### 7. Deep links and shareable URLs

**Problem:** Power users share links (e.g. “my focus list”, “this week’s review”) or want bookmarks and browser history to reflect exact view.

**Suggestion:**

- Ensure list view and search are fully encoded in the URL (e.g. `/list/:listId`, `/search?q=...`).
- Document URL format so users can construct links (e.g. for dashboards or internal wikis).
- Optional: “Copy link to this view” in the UI.

**Impact:** Integrates with how technical users share and bookmark.

---

### 8. Hide-at-times expression reference

**Problem:** Hide-at-times uses expr; power users want to write expressions without guessing the language.

**Suggestion:**

- Add a short “Hide-at-times” doc (or section under Options/TPPs) that:
  - States the environment (e.g. `D`, `H`, `M`) and points to [expr](https://github.com/expr-lang/expr) for syntax.
  - Gives 3–5 example expressions (e.g. “weekdays only”, “before 6pm”, “weekend”).
  - Mentions that RuleTest in the API can be used to validate expressions from scripts.

**Impact:** Makes a powerful feature discoverable and scriptable.

---

### 9. Quick capture from outside the app

**Problem:** Ideas and tasks happen in the terminal, browser, or other apps; opening the full UI is friction.

**Suggestion:**

- **Browser extension:** “Add to WackyTracky” that sends the current URL or selected text to the user’s instance (using the API) as a new task.
- **Global shortcut / launcher:** Document how to use a system shortcut or launcher (e.g. Raycast, Alfred, custom script) that calls the API to add a task; optionally provide a minimal “quick add” HTML page or CLI that only does “add one task” and is easy to host or open.

**Impact:** Captures tasks where they arise, without context-switching.

---

### 10. Templates and quick-add snippets

**Problem:** Power users repeat similar tasks (e.g. “weekly review”, “meeting with X”) and don’t want to retype tags/context every time.

**Suggestion:**

- **Templates:** User-defined templates (title + default content, e.g. “Meeting” with `#meeting @work`). “New from template” in add-task flow or a template dropdown.
- **Snippets:** Optional text snippets (e.g. type `;meeting` and expand to a standard meeting task line). Can be client-side only at first.

**Impact:** Speeds up data entry for repetitive workflows.

---

### 11. Advanced search (structured filters)

**Problem:** Power users often think in terms of “due before X”, “priority A”, “has subtasks”, “wait until past”.

**Suggestion:**

- If the backend supports it, extend search (and docs) with optional structured terms, e.g.:
  - `due:before:2025-03-01`, `due:after:today`, `priority:A`, `has:subtask`, `wait:over`.
- Keep the existing freetext and todo.txt-style terms; add these as an optional layer so power users can combine them.

**Impact:** More precise views and saved searches without leaving the app.

---

### 12. Git workflow documentation

**Problem:** Repo status is in the UI, but technical users want a clear workflow: when to commit, how to push, how to use branches if at all.

**Suggestion:**

- Document in the docs:
  - Where the todo.txt directory lives and how it’s used by the app.
  - Suggested workflow: e.g. “commit after a session” or “commit on done”; how to run `git push` (e.g. from host or cron).
  - Optional: UI button “Copy git commands” that suggests `git add ... && git commit -m "..."` (or link to a script) so users can run it in their own terminal.

**Impact:** Makes the “git integration” story concrete and repeatable.

---

### 13. Performance and scale

**Problem:** Power users with hundreds or thousands of tasks may hit slow renders or heavy requests.

**Suggestion:**

- Prefer list virtualization for long task lists so only visible rows are rendered.
- Lazy-load or paginate subtrees if a list has very deep or wide hierarchies.
- Keep list and search responses bounded (e.g. limit + “load more” or backend limits) and document any limits in the API docs.

**Impact:** The app stays fast and predictable at scale.

---

### 14. No artificial API limits

**Problem:** Many SaaS task tools rate-limit or restrict API access. Power users expect their self-hosted instance to be under their control.

**Suggestion:**

- Do not add per-user or per-IP rate limiting for the API beyond what’s needed for abuse prevention on a public instance. Document that self-hosted users are limited only by their own hardware and configuration.
- If optional rate limiting is ever added (e.g. for a public demo), make it configurable and default-off for self-hosted.

**Impact:** Reinforces that the app is built for people who run it themselves.

---

### 15. Webhooks or event hooks (optional)

**Problem:** Power users want to trigger CI, notifications, or other tools when tasks or lists change.

**Suggestion:**

- Optional webhook or “event hook” configuration (e.g. in config or UI): on events like “task done” or “task created”, POST a payload to a user-defined URL. Keep it simple (e.g. one URL per event type, optional secret for signing). Document payload format and idempotency so consumers can implement safely.

**Impact:** Enables integrations without polling the API.

---

## Summary table

| Suggestion                         | Effort (rough) | Impact for power users |
|------------------------------------|----------------|-------------------------|
| API documentation                  | Low            | High                   |
| CLI (or documented script)          | Medium         | High                   |
| Bulk operations                    | Medium         | High                   |
| Search syntax docs                 | Low            | High                   |
| Customizable shortcuts             | Medium         | Medium                 |
| Export/backup                      | Low–Medium     | High                   |
| Deep links / shareable URLs        | Low            | Medium                 |
| Hide-at-times reference            | Low            | Medium                 |
| Quick capture (extension/script)   | Medium         | Medium                 |
| Templates / snippets               | Medium         | Medium                 |
| Advanced search filters            | Medium         | Medium–High            |
| Git workflow docs                  | Low            | Medium                 |
| Performance (virtualization, etc.) | Medium         | High at scale          |
| No artificial API limits           | Low            | Trust / positioning    |
| Webhooks / event hooks             | Medium         | Medium                 |

---

*This list is a living set of suggestions. Implementation should follow the project’s specs (e.g. [main-task-list-ui](../specs/main-task-list-ui.md), [DESIGN](../specs/DESIGN.md)) and coding standards.*
