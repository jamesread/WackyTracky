# User-facing documentation (docs/)

This spec defines the purpose and scope of the **docs/** directory so that documentation stays focused on people who run and configure the software, not on contributors or implementers.

---

## Purpose

The docs in **docs/** are **user-facing**: usage instructions and configuration guidance for someone who wants to **run and use** the application (e.g. self-hosting). They are the main place for:

- How to install, run, and configure the application
- How to use the application day to day
- Configuration options, environment variables, and deployment (e.g. Docker) in terms of what users need to set and why

---

## Audience

- **Primary:** Entry-level self-hosters—people who may be new to this app but have a **working understanding of Docker** (running containers, volumes, basic networking). Do not assume deeper DevOps or software-development experience.
- **Out of scope:** Developer or contributor docs (how the codebase is structured, how to build from source, how to contribute). Those belong elsewhere (e.g. README, CONTRIBUTING, or internal wikis), not in user-facing docs.

---

## What to include

- **Usage:** How to accomplish common tasks in the app (from the user’s perspective).
- **Configuration:** What can be configured, what each option does, and what values are valid. Explain in terms of *behavior* and *outcomes*, not internal implementation.
- **Deployment:** How to run the app (e.g. with Docker/docker-compose), including any required env vars, volumes, or ports. Assume the reader is comfortable with basic Docker concepts.

---

## What to exclude

Do **not** include implementation or architecture details that are irrelevant to using and configuring the software. In particular, omit:

- How the **frontend** (Vue, Vite, components) is built or structured
- How the **service** (Go, APIs, internals) is implemented
- How the **protocol** (protobuf, connectrpc, etc.) is defined or used
- Build pipelines, test suites, or developer tooling

If a configuration option or behavior *must* be explained, describe it in terms of **what the user sees or needs**, not in terms of frontend/service/protocol layers.

---

## Scope

- **Applies to:** All content under **docs/** (e.g. MkDocs `source/` and any other user-facing doc assets).
- **When adding or editing docs:** Keep the audience (entry-level self-hosters with Docker) and the “usage + configuration only” rule in mind. Avoid slipping in developer-only or implementation-only sections unless they are clearly separated (e.g. an optional “Advanced” section that still stays user-oriented).

---

## Rationale

- **Single audience:** User docs that mix in implementation details become noisy and confusing for people who only want to run and configure the app.
- **Maintainability:** A clear boundary (user vs. developer) makes it easier to decide where new content belongs and to keep docs consistent.
- **Docker assumption:** Many self-hosted setups use containers; assuming basic Docker knowledge keeps docs concise without re-teaching Docker itself.
