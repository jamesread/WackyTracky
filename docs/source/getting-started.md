# Getting started

## Build

From the repository root:

```bash
make
```

This runs `generate`, then builds `service` and `frontend`. To build only the docs:

```bash
make docs
```

## Run

Start the backend from `service/` and the frontend dev server from `frontend/` according to the respective README or Makefile targets.

## Documentation (mkdocs)

To build the documentation site locally:

```bash
make -C docs
```

To install the docs build environment (mkdocs-material):

```bash
make -C docs devenv
```

The built site is in `docs/site/`. Serve it locally with `mkdocs serve` from the `docs/` directory if you have mkdocs installed.
