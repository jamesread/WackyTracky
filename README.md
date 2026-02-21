<div align = "center">
  <img alt = "project logo" src = "var/logo.png" width = "128" />
  <h1>WackyTracky</h1>

  WackyTracky gives **simple** task and list tracking with a minimal, keyboard-friendly interface.

[![Maturity](https://img.shields.io/badge/maturity-Beta-orange)](#none)
[![Discord](https://img.shields.io/discord/846737624960860180?label=Discord%20Server)](https://discord.gg/jhYWWpNJ3v)

</div>

**AI autonomy:** This project is developed at [Level 3 (semi-autonomous)](https://blog.jread.com/posts/ai-levels-of-autonomy-in-software-engineering/): AI implements features and tests; humans define specs and direction. Specs and architecture are written by a human with 20 years of development experience. Aggressive non-AI static analysis, code-quality checks (e.g. cyclomatic complexity limits), and tests are in place to reinforce project quality.

## Documentation

Full documentation (MkDocs) is published to **GitHub Pages**: [https://jamesread.github.io/WackyTracky/](https://jamesread.github.io/WackyTracky/). The site is built from `docs/` and deployed automatically on push to `main` (see `.github/workflows/docs.yml`).

## Deploy with Docker Compose

Container images are published to GitHub Container Registry. Quick reference to run with Docker Compose:

```yaml
# docker-compose.yml
services:
  wackytracky:
    image: ghcr.io/jamesread/wacky-tracky:latest
    ports:
      - "8080:8080"
    volumes:
      # Config: mount a file at /app/config.yaml or a directory at /config (with config.yaml inside)
      - ./config.yaml:/app/config.yaml
      # Persist todotxt data (set database.database in config to /app/data/todotxt or similar)
      - todotxt-data:/app/data
volumes:
  todotxt-data:
```

Then run: `docker compose up -d`. The web UI and API are served on port 8080 by default (adjust the image’s listen port in your config if needed). Place a `config.yaml` in the same directory as your `docker-compose.yml` if you mount it; see the [documentation](https://jamesread.github.io/WackyTracky/) for config options.

## Misc

* **Backend:** The service supports multiple database drivers (MySQL, Neo4j, YAML files, and [todo.txt](http://todotxt.org)) for different task systems. The **todotxt** driver is the one that by far receives the most focus and testing at the moment.
