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
      - wt-home:/home/wt
    restart: unless-stopped

volumes:
  todotxt-data:
  wt-home:
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
- **Home directory** — The image sets `HOME=/home/wt` and declares a volume at `/home/wt`. Mount it (as in the example above) to persist SSH keys for git push/pull from the container. Place private keys in `/home/wt/.ssh/` with mode `600` and the directory mode `700`. The image includes `git` and `openssh-clients`.
- **Port** — The app listens on port 8080 inside the container. Change the host port in the `ports` mapping if you want to use something other than 8080.

### Git over SSH

To use **Repo sync** or other git operations against a private remote:

1. Mount the home volume (see above), or bind-mount a host directory: `./wt-home:/home/wt`.
2. Copy your deploy key into the volume, e.g. `wt-home/.ssh/id_ed25519` (mode `600`).
3. Add a `wt-home/.ssh/config` if needed, for example:

```sshconfig
Host github.com
  IdentityFile ~/.ssh/id_ed25519
  IdentitiesOnly yes
```

4. Add the remote host to `wt-home/.ssh/known_hosts` (e.g. `ssh-keyscan github.com >> wt-home/.ssh/known_hosts` on the host before starting the container).

Your todo.txt data directory (`/app/data`) must still be a git repository with an upstream configured.

## Image and updates

- **Image:** `ghcr.io/jamesread/wacky-tracky:latest` (or pin a specific version tag).
- To update: pull the new image and recreate the container, e.g. `docker compose pull && docker compose up -d`. Your config and data volumes are unchanged.
