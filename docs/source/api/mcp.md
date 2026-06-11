# MCP server (for LLM assistants)

WackyTracky can run as an [MCP](https://modelcontextprotocol.io) server, letting an LLM assistant (for example Claude Desktop or Cursor) read and manage your tasks for you — "add "buy milk" to my shopping list", "what's on my work list?", "mark task X done".

The MCP server runs over **stdio**: your LLM client launches the WackyTracky binary as a subprocess. It uses the **same backend and configuration** as the normal server, so it reads and writes the same tasks (e.g. your todo.txt files).

## Running it

```bash
wt mcp
```

This starts the server on standard input/output and waits for an MCP client to connect. You normally don't run this by hand — your LLM client launches it for you (see below). It reads `config.yaml` from the usual locations (current directory, `../`, or `/config`).

## Configuring your LLM client

Point your client at the WackyTracky binary with the `mcp` argument. Most clients use a JSON config like this:

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

If the data directory for the todo.txt backend is not discoverable from where the client launches the process, set it via environment or run from a directory that contains your `config.yaml`:

```json
{
  "mcpServers": {
    "wacky-tracky": {
      "command": "/usr/local/bin/wt",
      "args": ["mcp"],
      "env": { "DATABASE_DRIVER": "todotxt", "DATABASE_DATABASE": "/path/to/your/todo-dir" }
    }
  }
}
```

The exact location of this config differs per client (e.g. Claude Desktop's `claude_desktop_config.json`, or a `.cursor/mcp.json` for Cursor).

## Available tools

| Tool | Description |
|------|-------------|
| `list_lists` | List all task lists with their IDs, titles, and item counts. |
| `list_tasks` | List the tasks in a list (including subtasks). Requires a `list_id`. |
| `search_tasks` | Search tasks across all lists (`#tag`, `@context`, `-term` to exclude). |
| `create_task` | Create a task (todo.txt-style content; optional `list_id`/`parent_task_id`). |
| `update_task` | Replace a task's content. |
| `complete_task` | Mark a task as done. |
| `create_list` | Create a new list. |

The tools call the same operations as the [HTTP API](index.md), so behaviour (hide-at-times filtering, todo.txt parsing of `#tags`/`@contexts`, etc.) is identical.

## Security note

The MCP server has the same access to your tasks as the WackyTracky server itself, with no extra authentication — anything that can launch `wt mcp` can read and modify your tasks. Only configure it for LLM clients you trust, and treat it like giving that client direct access to your task data.
