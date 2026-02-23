# Service API

The WackyTracky backend exposes a **Connect RPC** API. The web app uses this API for all operations (lists, tasks, search, metadata, and more). You can call it from scripts, other applications, or a CLI using HTTP and JSON.

## Base URL

When the server is running (e.g. on port 8080), the API base URL is:

```
http://<host>:<port>/api
```

Example: `http://localhost:8080/api`

Each RPC has a procedure path under this base. The full URL for a procedure is:

```
<base>/wackytracky.clientapi.v1.WackyTrackyClientService/<MethodName>
```

Example: `http://localhost:8080/api/wackytracky.clientapi.v1.WackyTrackyClientService/GetLists`

## Protocol

The service uses the **Connect** protocol (HTTP/JSON):

- **Method:** `POST`
- **Request:** `Content-Type: application/json`, body is a JSON object with the request message fields (empty `{}` for empty requests).
- **Response:** `Content-Type: application/json`, body is a JSON object with the response message fields.
- **Errors:** On failure, the server returns an appropriate HTTP status (e.g. 4xx/5xx) and a Connect error payload in JSON.

If your deployment uses HTTP authentication (e.g. the optional [httpauthshim](https://github.com/jamesread/httpauthshim) integration), include credentials in your requests (e.g. `Authorization: Basic ...` for Basic auth).

## Quick start

### Version (no arguments)

```bash
curl -s -X POST 'http://localhost:8080/api/wackytracky.clientapi.v1.WackyTrackyClientService/Version' \
  -H "Content-Type: application/json" \
  -d '{}'
```

Example response:

```json
{"version":"1.9.0","commit":"abc123","date":"2025-01-15T12:00:00Z"}
```

### Get lists

```bash
curl -s -X POST 'http://localhost:8080/api/wackytracky.clientapi.v1.WackyTrackyClientService/GetLists' \
  -H "Content-Type: application/json" \
  -d '{}'
```

Example response:

```json
{"lists":[{"id":"list-id-1","title":"Inbox","countItems":5}]}
```

### Create a task

```bash
curl -s -X POST 'http://localhost:8080/api/wackytracky.clientapi.v1.WackyTrackyClientService/CreateTask' \
  -H "Content-Type: application/json" \
  -d '{"content":"Buy milk #errands @home","parentListId":"<list-id>"}'
```

Use the list ID from `GetLists`. Omit `parentListId` to use the default list (implementation-dependent). Use `parentTaskId` to create a subtask.

### Search tasks

```bash
curl -s -X POST 'http://localhost:8080/api/wackytracky.clientapi.v1.WackyTrackyClientService/SearchTasks' \
  -H "Content-Type: application/json" \
  -d '{"query":"#work -#done"}'
```

Query syntax is backend-dependent; the todo.txt backend supports plain terms, `#tag`, `@context`, and `-term` to exclude. See [Search (SearchTasks)](search.md) for details.

### Mark a task done

```bash
curl -s -X POST 'http://localhost:8080/api/wackytracky.clientapi.v1.WackyTrackyClientService/DoneTask' \
  -H "Content-Type: application/json" \
  -d '{"id":"<task-id>"}'
```

Task IDs come from `ListTasks` or `SearchTasks` responses.

## Reference

The full list of RPCs and request/response message types is in the [API reference](reference.md).

## Code generation

The API is defined with Protocol Buffers in `protocol/wacky-tracky/clientapi/v1/wt.proto`. You can generate client code using:

- **Connect:** [connectrpc.com](https://connectrpc.com) (Go, TypeScript/JavaScript, etc.)
- **buf:** [buf.build](https://buf.build) with the Connect plugins

The web frontend uses the generated Connect-Web client; the Go server implements the same service.
