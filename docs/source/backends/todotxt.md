# todo.txt backend

The **todo.txt** backend is the **default** and **recommended** storage option for WackyTracky. It is the backend that receives **most development and testing**.

Your tasks and lists are stored as plain text files in the [todo.txt](http://todotxt.org) format. You can edit these files with any text editor, sync them with Dropbox or Git, and use other todo.txt-compatible tools.

## Configuration

Set the backend and data directory in your config (e.g. `config.yaml`):

```yaml
database:
  driver: todotxt
  database: /path/to/your/todotxt/directory
```

- **driver:** `todotxt`
- **database:** Path to the directory where todo.txt files will be stored (e.g. `todo.txt`, `done.txt`, and list metadata). The directory will be created if it does not exist.

## With Docker

When running in Docker, use a volume or bind mount for this directory and set `database.database` to the path inside the container (e.g. `/app/data`). See [Installation — Docker Compose](../installation/docker-compose.md).

## Data format

Tasks are stored in `todo.txt` (incomplete) and `done.txt` (completed). The format is compatible with the [todo.txt](http://todotxt.org) spec, so you can use the same files in other apps or scripts.
