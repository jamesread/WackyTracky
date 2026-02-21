# Docker Compose

The easiest way to run WackyTracky is with Docker Compose. Container images are published to GitHub Container Registry.

## Prerequisites

- Docker and Docker Compose installed
- A directory for your compose file and (optionally) config

## Quick start

1. Create a directory for your deployment and add a `docker-compose.yml`:

```yaml
services:
  wackytracky:
    image: ghcr.io/jamesread/wacky-tracky:latest
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/app/config.yaml
      - todotxt-data:/app/data
    restart: unless-stopped

volumes:
  todotxt-data:
```

2. Add a `config.yaml` in the same directory (next to `docker-compose.yml`). For the default **todo.txt** backend, you can use:

```yaml
database:
  driver: todotxt
  database: /app/data
```

3. Start the stack:

```bash
docker compose up -d
```

4. Open the web UI at **http://localhost:8080** (or your host’s IP if running on a server).

## Configuration

- **Config file** — You can mount a single file at `/app/config.yaml` (as above) or a directory at `/config` that contains `config.yaml`. The app looks for `config.yaml` in `/config` when running in the container.
- **Data** — For the todo.txt backend, persist the data directory (e.g. `todotxt-data` volume or a bind mount) and set `database.database` to that path (e.g. `/app/data`). The image already uses `/app/data` by default if you don’t override it.
- **Port** — The app listens on port 8080 inside the container. Change the host port in the `ports` mapping if you want to use something other than 8080.

## Image and updates

- **Image:** `ghcr.io/jamesread/wacky-tracky:latest` (or pin a specific version tag).
- To update: pull the new image and recreate the container, e.g. `docker compose pull && docker compose up -d`. Your config and data volumes are unchanged.
