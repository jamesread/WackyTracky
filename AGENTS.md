# AI & agent integration

WackyTracky exposes machine-readable discovery endpoints on the same base URL as the web app:

| Surface | Path | Purpose |
|---------|------|---------|
| `llms.txt` | `/llms.txt` | Human/LLM-readable index of integration options |
| OpenAPI | `/openapi` | OpenAPI 3.1 YAML spec for the Connect RPC API |
| MCP | `/mcp` | Model Context Protocol (Streamable HTTP) for AI assistants |

## Authentication

IAM (users, sessions, API keys, RBAC) uses a **separate SQLite** database (`iam.database` in config). The Connect API and HTTP MCP share armature-iam / httpauthshim auth: browser cookie `wacky-tracky-sid`, or `Authorization: Bearer` API keys (`wt_` prefix). First empty IAM DB seeds `admin` / `admin`. Dev: `WACKYTRACKY_DEV_DISABLE_AUTH=true`. See [docs/modules/ROOT/pages/authentication.adoc](docs/modules/ROOT/pages/authentication.adoc).

## Connect RPC API

Base path: `/api`

Every procedure is a POST with a JSON body to:

`/api/wackytracky.clientapi.v1.WackyTrackyClientService/<MethodName>`

Use the [OpenAPI spec](docs/modules/ROOT/attachments/openapi.yaml) (also at `GET /openapi` on a running server) for request/response schemas. For non-MCP integrations, prefer calling the Connect API directly or generating clients from OpenAPI.

## MCP server

**HTTP (recommended when the server is reachable):**

```json
{
  "mcpServers": {
    "wacky-tracky": {
      "url": "https://your-host/mcp"
    }
  }
}
```

Add `Authorization: Bearer <api-key>` for HTTP MCP when auth is enabled (cookie auth is not used on `/mcp`).

**Stdio (local subprocess):**

```json
{
  "mcpServers": {
    "wacky-tracky": {
      "command": "wt",
      "args": ["mcp"]
    }
  }
}
```

The stdio server uses the same backend and configuration as the HTTP server (`config.yaml`, environment variables).

### MCP tools

| Tool | Parameters | Description |
|------|------------|-------------|
| `list_lists` | — | List all task lists with IDs, titles, and item counts |
| `list_tasks` | `list_id` (required) | List tasks in a list, including subtasks |
| `search_tasks` | `query` (required) | Search across lists (`#tag`, `@context`, `-term` to exclude) |
| `create_task` | `content` (required), `list_id`, `parent_task_id` | Create a todo.txt-style task |
| `update_task` | `id`, `content` (required) | Replace a task's content |
| `complete_task` | `id` (required) | Mark a task done |
| `create_list` | `title` (required) | Create a new list |

Tools return JSON (camelCase fields) matching the Connect API responses.

---

# Commits (hard requirement)

- **Use [Conventional Commits](https://www.conventionalcommits.org/)** for every commit. Commit messages must follow the format: `type(scope): description` (e.g. `feat(api): add task search`, `fix(frontend): toast on save error`). Types include but are not limited to: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `ci`, `chore`. This is enforced by the commit-msg hook (conventional-pre-commit); the build and release pipeline depend on it.

# Coding style

- Use comments to explain edge cases or design decisions only. Don't use them to describe what the code is doing.
- Use the languages' standard style guides for formatting and naming conventions.
- Make function and method names descriptive of their behavior.
- Cyclomic complexity should be kept to 4 or less.

# TDD

- Write tests for all new features and bug fixes.
- If a test breaks due to a code change, prompt the user to decide whether to update the test or fix the code.

# Project layout

- Make and makefile targets should be used to build, test, lint, and run the project. Make files should call other Makefiles for subdirectories like `make -wC frontend`.
- Running `make` without arguments should build the project and all subdirectories.
- Backend code should be in the service/ directory.
- Frontend code (js, vue, vite, etc.) should be in the frontend/ directory.
- Integration tests should be in the integration-tests/ directory.
- Protocol code should use connectrpc and buf, and should be in the protocol/ directory.
- Documentation should be in the docs/ directory, and use Antora.

# Specs (codebase-specific guidance)

- The **specs/** directory contains design decisions and implementation guidance for this codebase. When working on the frontend, on features that touch UX, testing, or dialogs, or on the **docs/** directory, consult the specs and follow them unless the user instructs otherwise.
- Specs include: Vue-based dialogs (no window.prompt/confirm/alert), toast feedback for async operations, main task list minimalism, **Go development** (cyclo ≤ 5, codestyle and unit tests must pass; see `specs/go-development.md`), **user-facing docs** (usage and configuration for entry-level self-hosters; see `specs/docs-user-facing.md`), and other product/design decisions. Treat specs as normative for implementation and code review where they apply.

# repo health

- The command line tool `repohealth` reports issues with the repo.

# Frontend

- Use a component-based architecture with Vue + Vite.
- For routing, use Vue Router.
- Use the npm library `picocrank` that provides some common components and utilities.
- Use as few CSS rules as possible in the components, the library femtocrank which is a dependency of picocrank provides most of the needed styling.

# Backend

- Use Go modules for dependency management.
- Use the standard library as much as possible.
- Use the `logrus` library for logging.
- Use the `koanf` library for configuration management (YAML `config.yaml`).
- Listen address: honour `$PORT` when set (via `golure/pkg/listenaddr`), otherwise `listenAddress` from config (default `0.0.0.0:8080`). Vite’s `/api` proxy uses `$PORT` or `8080`.
- Use the `jamesread/golure` library for utility functions.
- Use the `jamesread/httpauthshim` library for HTTP authentication.
- Use the `stretchr/testify` library for testing.
- Use the `connectrpc` library for gRPC services.

# Protocol

- Use Protocol Buffers version 3 syntax.
- Use `connectrpc` for gRPC services.
- Use `buf` for managing Protocol Buffers files and generating code.
- Follow best practices for designing Protocol Buffers messages and services.
- Keep Protocol Buffers files organized and modular.
- Document Protocol Buffers messages and services with comments.

# Testing

- Use unit tests for individual functions and methods.
- Use integration tests for testing interactions between components.
- Use end-to-end tests for testing the entire system.
- Use mocking and stubbing to isolate components during testing.
- integration tests should be implemented using mocha and selenium-webdriver.
- integration tests should be located in the integration-tests/tests/ directory, and include the JS tests, and config.yaml for the backend.
- integration tests should start and stop the backend service and set the -configdir arg as needed.
