# Go code standards

This spec defines requirements for Go code in this repo so that the service stays maintainable and changes are validated before commit.

---

## Principle: low complexity and passing checks

Go code in `service/` must:

1. **Keep cyclomatic complexity under 5** per function. Complex functions are split into smaller helpers so that no function exceeds a complexity of 5 (as measured by `gocyclo -over 5`).
2. **Pass codestyle and unit tests** after any code change. Before committing, `go fmt`, `go vet`, and `go test` must succeed for the Go module in `service/`.

These requirements are enforced by pre-commit: the hooks run when any `service/**/*.go` file is staged. Commits that would leave the tree with complexity &gt; 5 or failing tests/codestyle are blocked.

---

## Cyclomatic complexity (cyclo ≤ 5)

- **Tool:** `gocyclo -over 5 $(find internal -name '*.go' ! -name '*_test.go')` (run from `service/`). No *production* function in `internal/` may have complexity greater than 5; `*_test.go` files are excluded so test helpers with many assertions are allowed.
- **Refactor approach:** Extract helpers, reduce branching, and use early returns so that each function stays small and testable. Prefer many small functions over a few large ones.
- **Pre-commit:** A pre-commit hook runs gocyclo over production `.go` files in `internal/` (excluding `*_test.go`) when any `service/**/*.go` file is staged and fails if any function is over the limit. Install the tool with `make go-tools` (or `go install github.com/fzipp/gocyclo/cmd/gocyclo@latest`).

---

## Codestyle and tests

- **Format:** Code must be formatted with `go fmt ./...` (run from `service/`). No unformatted Go code may be committed.
- **Static checks:** `go vet ./...` must pass. Fix any reported issues before committing.
- **Unit tests:** `go test ./...` must pass. New or changed behavior should be covered by unit tests where practical; at least existing tests must not be broken.

Pre-commit runs `go fmt`, `go vet`, and `go test` for `service/` when Go files are staged. Ensure these pass locally (e.g. `cd service && go fmt ./... && go vet ./... && go test ./...`) before pushing.

---

## Scope

- **Applies to:** All Go code under `service/`, in particular `service/internal/`.
- **When changing Go code:** Re-run `gocyclo -over 5 internal/`, `go fmt ./...`, `go vet ./...`, and `go test ./...` from `service/` after making changes. Fix any failures before committing.

---

## Rationale

- **Complexity:** Low cyclomatic complexity keeps functions easier to read, test, and change. A hard limit (e.g. 5) prevents accidental complexity creep and encourages refactors into smaller units.
- **Checks before commit:** Catching format, vet, and test failures at pre-commit keeps the main branch green and avoids “fix the build” commits. Developers get immediate feedback when their changes break style or tests.
