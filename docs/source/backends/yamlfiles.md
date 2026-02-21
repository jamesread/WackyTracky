# YAML files backend

The **YAML files** backend stores tasks, tags, and lists in YAML files on disk (`tasks.yaml`, `tags.yaml`, `lists.yaml`). Use this if you want a simple file-based store that is not todo.txt format.

This backend is **not** the default and receives less development focus than the todo.txt backend.

## Configuration

In your config (e.g. `config.yaml`):

```yaml
database:
  driver: yamlfiles
  database: /path/to/directory/containing/yaml/files
```

- **driver:** `yamlfiles`
- **database:** Path to the directory where the YAML files will be read and written.

The backend expects or creates `tasks.yaml`, `tags.yaml`, and `lists.yaml` in that directory.

## With Docker

Mount a volume or bind mount for the data directory and set `database.database` to the path inside the container so the app can read and write the YAML files.

## When to use

Choose YAML files if you prefer YAML over todo.txt and don’t need Neo4j. For most users, the [todo.txt](todotxt.md) backend is recommended and better supported.
