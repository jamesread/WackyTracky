# API reference

Service: **WackyTrackyClientService**
Package: `wackytracky.clientapi.v1`

Base path for all procedures:

```
POST /api/wackytracky.clientapi.v1.WackyTrackyClientService/<MethodName>
```

Request and response bodies are JSON with the field names as in the tables below (camelCase in JSON: e.g. `parent_list_id` → `parentListId` for compatibility with Connect/Protobuf JSON mapping).

---

## RPCs

### Version

Returns server version and build info.

| Procedure | `Version` |
|-----------|-----------|
| **Request** | `VersionRequest` (empty) |
| **Response** | `VersionResponse` |

**VersionResponse**

| Field    | Type   | Description        |
|----------|--------|--------------------|
| version  | string | Version string     |
| commit   | string | Git commit         |
| date     | string | Build date (ISO8601) |

---

### Init

Returns app init data (e.g. wallpaper path).

| Procedure | `Init` |
|-----------|--------|
| **Request** | `InitRequest` (empty) |
| **Response** | `InitResponse` |

**InitResponse**

| Field     | Type   | Description |
|-----------|--------|-------------|
| wallpaper | string | Wallpaper path or identifier |

---

### GetLists

Returns all lists.

| Procedure | `GetLists` |
|-----------|------------|
| **Request** | `GetListsRequest` (empty) |
| **Response** | `GetListsResponse` |

**GetListsResponse**

| Field  | Type    | Description   |
|--------|---------|---------------|
| lists  | List[]  | All lists     |

---

### ListTasks

Returns tasks for a list or for a parent task (children). Used to load a list view or a subtree.

| Procedure | `ListTasks` |
|-----------|-------------|
| **Request** | `ListTasksRequest` |
| **Response** | `ListTasksResponse` |

**ListTasksRequest**

| Field       | Type   | Description |
|-------------|--------|-------------|
| parent_id   | string | List ID (for root tasks) or task ID (for subtasks) |
| parent_type | string | `"list"` or `"task"` |

**ListTasksResponse**

| Field                 | Type              | Description |
|-----------------------|-------------------|-------------|
| tasks                 | Task[]            | Tasks for the requested parent |
| tree                  | map<string, TaskIdList> | Parent ID → ordered child task IDs (for hierarchy) |
| hidden_tag_names      | string[]          | Tags currently hidden by hide-at-times rules |
| hidden_context_names  | string[]          | Contexts currently hidden by hide-at-times rules |

**TaskIdList**

| Field | Type     | Description |
|-------|----------|-------------|
| ids   | string[] | Ordered task IDs |

---

### SearchTasks

Searches tasks across lists. Query syntax is backend-dependent (todo.txt: plain terms, `#tag`, `@context`, `-term` to exclude).

| Procedure | `SearchTasks` |
|-----------|---------------|
| **Request** | `SearchTasksRequest` |
| **Response** | `SearchTasksResponse` |

**SearchTasksRequest**

| Field | Type   | Description |
|-------|--------|-------------|
| query | string | Search query (freetext; backends may support structured syntax) |

**SearchTasksResponse**

| Field  | Type   | Description   |
|--------|--------|---------------|
| tasks  | Task[] | Matching tasks |

---

### CreateTask

Creates a new task, optionally in a list or as a subtask.

| Procedure | `CreateTask` |
|-----------|--------------|
| **Request** | `CreateTaskRequest` |
| **Response** | `CreateTaskResponse` |

**CreateTaskRequest**

| Field           | Type   | Description |
|-----------------|--------|-------------|
| content         | string | Task content (todo.txt line: text, +projects, @contexts, #tags) |
| parent_list_id  | string | Optional. List ID to add the task to |
| parent_task_id  | string | Optional. Task ID to create this as a subtask of |

**CreateTaskResponse**

| Field  | Type | Description   |
|--------|------|---------------|
| task   | Task | Created task  |

---

### UpdateTask

Updates a task’s content.

| Procedure | `UpdateTask` |
|-----------|--------------|
| **Request** | `UpdateTaskRequest` |
| **Response** | `UpdateTaskResponse` |

**UpdateTaskRequest**

| Field   | Type   | Description |
|---------|--------|-------------|
| id      | string | Task ID     |
| content | string | New content (full todo.txt-style line) |

**UpdateTaskResponse**

| Field  | Type | Description   |
|--------|------|---------------|
| task   | Task | Updated task  |

---

### DoneTask

Marks a task as done (moves to done.txt in todo.txt backend).

| Procedure | `DoneTask` |
|-----------|------------|
| **Request** | `DoneTaskRequest` |
| **Response** | `DoneTaskResponse` (empty) |

**DoneTaskRequest**

| Field | Type   | Description   |
|-------|--------|----------------|
| id    | string | Task ID to mark done |

---

### MoveTask

Moves a task to another list.

| Procedure | `MoveTask` |
|-----------|------------|
| **Request** | `MoveTaskRequest` |
| **Response** | `MoveTaskResponse` (empty) |

**MoveTaskRequest**

| Field           | Type   | Description   |
|-----------------|--------|---------------|
| task_id         | string | Task ID       |
| target_list_id  | string | Destination list ID |

---

### CreateList

Creates a new list.

| Procedure | `CreateList` |
|-----------|--------------|
| **Request** | `CreateListRequest` |
| **Response** | `CreateListResponse` (empty) |

**CreateListRequest**

| Field  | Type   | Description |
|--------|--------|-------------|
| title  | string | List title   |

---

### UpdateList

Updates a list’s title.

| Procedure | `UpdateList` |
|-----------|--------------|
| **Request** | `UpdateListRequest` |
| **Response** | `UpdateListResponse` (empty) |

**UpdateListRequest**

| Field  | Type   | Description |
|--------|--------|-------------|
| id     | string | List ID     |
| title  | string | New title   |

---

### DeleteList

Deletes a list.

| Procedure | `DeleteList` |
|-----------|--------------|
| **Request** | `DeleteListRequest` |
| **Response** | `DeleteListResponse` (empty) |

**DeleteListRequest**

| Field | Type   | Description |
|-------|--------|-------------|
| id    | string | List ID     |

---

### GetTags

Returns all tags (for UI/autocomplete). Behavior may vary by backend.

| Procedure | `GetTags` |
|-----------|-----------|
| **Request** | `GetTagsRequest` (empty) |
| **Response** | `GetTagsResponse` |

**GetTagsResponse**

| Field | Type   | Description |
|-------|--------|-------------|
| tags  | Tag[]  | Tags        |

**Tag**

| Field  | Type   | Description |
|--------|--------|-------------|
| id     | string | Tag ID      |
| title  | string | Tag label   |

---

### RepoStatus

Returns `git status` output for the todo.txt data directory (todo.txt backend). Useful for power users who version their data with Git.

| Procedure | `RepoStatus` |
|-----------|--------------|
| **Request** | `RepoStatusRequest` (empty) |
| **Response** | `RepoStatusResponse` |

**RepoStatusResponse**

| Field   | Type   | Description |
|---------|--------|-------------|
| output  | string | Git status output, or error message if not a git repo / git unavailable |

---

### GetSavedSearches

Returns the user’s saved searches (names and queries).

| Procedure | `GetSavedSearches` |
|-----------|--------------------|
| **Request** | `GetSavedSearchesRequest` (empty) |
| **Response** | `GetSavedSearchesResponse` |

**GetSavedSearchesResponse**

| Field          | Type          | Description   |
|----------------|---------------|---------------|
| saved_searches | SavedSearch[] | Saved searches |

**SavedSearch**

| Field  | Type   | Description |
|--------|--------|-------------|
| id     | string | Unique ID   |
| name   | string | Display name |
| query  | string | Search query |

---

### SetSavedSearches

Replaces the user’s saved searches. Persistence is backend-dependent (e.g. stored in browser and/or synced to server).

| Procedure | `SetSavedSearches` |
|-----------|--------------------|
| **Request** | `SetSavedSearchesRequest` |
| **Response** | `SetSavedSearchesResponse` (empty) |

**SetSavedSearchesRequest**

| Field          | Type          | Description   |
|----------------|---------------|---------------|
| saved_searches | SavedSearch[] | Full list of saved searches to store |

---

### GetTaskMetadata

Returns metadata fields for a task (e.g. notes, priority). Optional per backend.

| Procedure | `GetTaskMetadata` |
|-----------|--------------------|
| **Request** | `GetTaskMetadataRequest` |
| **Response** | `GetTaskMetadataResponse` |

**GetTaskMetadataRequest**

| Field   | Type   | Description |
|---------|--------|-------------|
| task_id | string | Task ID     |

**GetTaskMetadataResponse**

| Field  | Type            | Description |
|--------|-----------------|-------------|
| fields | map<string, string> | Field name → value (e.g. `notes`, `priority`, `wait`, `due`) |

---

### SetTaskMetadata

Sets a single metadata field for a task (e.g. notes, priority, wait, due).

| Procedure | `SetTaskMetadata` |
|-----------|--------------------|
| **Request** | `SetTaskMetadataRequest` |
| **Response** | `SetTaskMetadataResponse` (empty) |

**SetTaskMetadataRequest**

| Field   | Type   | Description |
|---------|--------|-------------|
| task_id | string | Task ID     |
| field   | string | Field name (e.g. `notes`, `priority`, `wait`, `due`) |
| value   | string | Field value |

---

### GetTaskPropertyProperties

Returns Task Property Properties (TPPs): per-tag and per-context settings (e.g. CSS, hide-at-times rules).

| Procedure | `GetTaskPropertyProperties` |
|-----------|-----------------------------|
| **Request** | `GetTaskPropertyPropertiesRequest` (empty) |
| **Response** | `GetTaskPropertyPropertiesResponse` |

**GetTaskPropertyPropertiesResponse**

| Field               | Type                      | Description |
|---------------------|---------------------------|-------------|
| tag_properties      | map<string, TaskPropertyProps> | Tag name → props |
| context_properties  | map<string, TaskPropertyProps> | Context name → props |

**TaskPropertyProps**

| Field  | Type            | Description |
|--------|-----------------|-------------|
| props  | map<string, string> | Key → value. Supported keys: `css`, `hide-at-times` (expr expression) |

---

### SetTaskPropertyProperty

Sets one TPP key for a tag or context (e.g. `css` or `hide-at-times`).

| Procedure | `SetTaskPropertyProperty` |
|-----------|----------------------------|
| **Request** | `SetTaskPropertyPropertyRequest` |
| **Response** | `SetTaskPropertyPropertyResponse` (empty) |

**SetTaskPropertyPropertyRequest**

| Field          | Type   | Description |
|----------------|--------|-------------|
| property_type  | string | `"tag"` or `"context"` |
| property_name  | string | Tag or context name |
| key            | string | Prop key (e.g. `css`, `hide-at-times`) |
| value          | string | Prop value  |

---

### RuleStatus

Returns current date/time components for building hide-at-times expressions (D = weekday short name, H = hour, M = minute).

| Procedure | `RuleStatus` |
|-----------|--------------|
| **Request** | `RuleStatusRequest` (empty) |
| **Response** | `RuleStatusResponse` |

**RuleStatusResponse**

| Field | Type  | Description |
|-------|-------|-------------|
| D     | string | Short day: Mon, Tue, …, Sun |
| H     | int32  | Hour 0–23   |
| M     | int32  | Minute 0–59 |

---

### RuleTest

Compiles and evaluates a hide-at-times expression. Use to validate or preview rules (e.g. “hide @work on weekends”).

| Procedure | `RuleTest` |
|-----------|------------|
| **Request** | `RuleTestRequest` |
| **Response** | `RuleTestResponse` |

**RuleTestRequest**

| Field      | Type   | Description |
|------------|--------|-------------|
| expression | string | Expr expression (can use D, H, M) |

**RuleTestResponse**

| Field         | Type   | Description |
|---------------|--------|-------------|
| compiles      | bool   | Whether the expression compiled |
| compile_error | string | Set when compiles is false |
| result        | bool   | Eval result when compiles is true |
| eval_error    | string | Set when compile ok but eval failed |

---

### Tag

Reserved RPC (Tag). Behavior is implementation-specific; may be used for tag-related actions.

| Procedure | `Tag` |
|-----------|-------|
| **Request** | `TagRequest` (empty) |
| **Response** | `TagResponse` (empty) |

---

## Shared message types

### List

| Field        | Type   | Description   |
|--------------|--------|---------------|
| id           | string | List ID       |
| title        | string | List title    |
| count_items  | int32  | Item count    |

### Task

| Field           | Type     | Description |
|-----------------|----------|-------------|
| id              | string   | Task ID     |
| content         | string   | Full content (todo.txt-style line) |
| parent_id       | string   | Parent list or task ID |
| parent_type     | string   | `"list"` or `"task"` |
| count_subitems  | int32    | Number of direct children |
| tags            | string[] | Parsed #tags |
| contexts        | string[] | Parsed @contexts |
| wait_until      | string   | Optional ISO date-time; task hidden until this time |
| priority        | string   | Optional single letter A–Z |
| due_date        | string   | Optional ISO date YYYY-MM-DD |
