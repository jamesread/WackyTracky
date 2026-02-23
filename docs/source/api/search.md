# Search (SearchTasks)

The **SearchTasks** RPC searches over incomplete tasks. Behavior is backend-dependent; the **todo.txt** backend is described here.

## Query format

- The query string is split on **whitespace** into terms.
- **Include terms:** A task is returned only if its searchable text contains every include term (case-insensitive substring).
- **Exclude terms:** A term that starts with `-` is an exclude term (e.g. `-home`). A task is excluded if its searchable text contains the rest of the term (e.g. `home`).

## Searchable text

For each task, the searchable text includes:

- The task’s description (body)
- Its **projects** (e.g. `+project`)
- Its **contexts** (e.g. `@home`, `@work`)
- Its **tags** (e.g. `#work`, `#errands`)
- The same from all **ancestor tasks** (so a subtask matches when a parent has the tag or context)

So you can search for:

- Plain words: `milk` matches tasks containing “milk”.
- Tags: `#work` matches tasks (or subtasks of tasks) that have the tag `#work`.
- Contexts: `@home` matches tasks with context `@home`.
- Exclude: `#work -@meeting` matches tasks tagged `#work` that do not contain `@meeting`.

Only **incomplete** tasks (in `todo.txt`) are searched; completed tasks (in `done.txt`) are not included.

## Example

```bash
curl -s -X POST 'http://localhost:8080/api/wackytracky.clientapi.v1.WackyTrackyClientService/SearchTasks' \
  -H "Content-Type: application/json" \
  -d '{"query":"#errands @home"}'
```

Returns all incomplete tasks that have both tag `#errands` and context `@home`.
