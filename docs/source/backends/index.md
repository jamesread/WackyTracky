# Backends

WackyTracky can store tasks and lists in different backends. You choose one when configuring the app.

## Recommended: todo.txt

**[todo.txt](todotxt.md)** is the **default** backend. It is **recommended** for most users and is the one that receives **most development and testing**. Your data lives as plain text files in the [todo.txt](http://todotxt.org) format, so you can edit, sync, or back up your tasks with standard tools.

Use todo.txt unless you have a specific reason to use another backend.

## Other backends

- **[Neo4j](neo4j.md)** — Store tasks in a Neo4j graph database. For users who already run Neo4j or want a graph-based store.
- **[YAML files](yamlfiles.md)** — Store tasks in YAML files on disk. A simple file-based option that does not use the todo.txt format.

Each backend has its own configuration options; see the linked pages for details.
