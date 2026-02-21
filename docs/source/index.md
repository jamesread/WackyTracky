# WackyTracky

WackyTracky is a minimal, keyboard-friendly task and list tracker. It gives you simple task and list tracking with a clean interface that works well for power users and GTD-style workflows.

## What you can do

- **Tasks and lists** — Organize tasks into lists, with support for nested tasks, tags, and contexts.
- **todo.txt compatibility** — The default storage uses the [todo.txt](http://todotxt.org) format, so your data stays in plain text files you can edit or sync with other tools.
- **Self-hosted** — Run it yourself with Docker; your data stays on your machine or server.

## Get started

1. **[Install with Docker Compose](installation/docker-compose.md)** — Recommended way to run WackyTracky. One `docker-compose.yml` and a small config file are all you need.
2. **Choose a backend** — WackyTracky can store tasks in different backends. **[todo.txt](backends/todotxt.md)** is the default and recommended option and is where most development and testing happens. Other options: [Neo4j](backends/neo4j.md), [YAML files](backends/yamlfiles.md).

## Documentation

- [Installation](installation/docker-compose.md) — Run with Docker Compose
- [Backends](backends/index.md) — Storage options (todo.txt, Neo4j, YAML files)
