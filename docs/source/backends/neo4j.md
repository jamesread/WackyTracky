# Neo4j backend

The **Neo4j** backend stores tasks and lists in a [Neo4j](https://neo4j.com/) graph database. Use this backend if you already run Neo4j or want a graph-based store.

This backend is **not** the default and receives less development focus than the todo.txt backend.

## Configuration

In your config (e.g. `config.yaml`):

```yaml
database:
  driver: neo4j
  hostname: localhost
  port: 7687
  username: neo4j
  password: your-password
```

- **driver:** `neo4j`
- **hostname:** Neo4j server hostname
- **port:** Bolt port (default 7687)
- **username:** Neo4j user
- **password:** Neo4j password

## With Docker

Run Neo4j in a separate container (or use an existing Neo4j instance). Ensure the WackyTracky container can reach Neo4j on the Bolt port. Use the service name or hostname of your Neo4j container as `hostname` if they are on the same Docker network.

## When to use

Choose Neo4j if you need graph storage or already have Neo4j in your stack. For most users, the [todo.txt](todotxt.md) backend is simpler and recommended.
