# WackyTracky integration tests

Integration tests for the WackyTracky web UI using Mocha and Selenium WebDriver.

## Prerequisites

- Node.js (for npm, mocha, selenium-webdriver)
- Chrome or Chromium (for headless browser)
- Go (service is built before tests)
- Frontend and service must be buildable from repo root (`make frontend`, `make service`)

## Run

From repo root:

```bash
make integration-tests
```

Or from this directory:

```bash
npm install --no-fund
npx mocha tests --recursive -t 15000
```

The runner builds the frontend and service, starts the server (with config from `configs/default/`), runs the tests, then stops the server.

## Structure

- `configs/default/config.yaml` — config used when starting the server (copied to service/ for the run).
- `lib/elements.js` — shared helpers (e.g. getRootAndWait, takeScreenshotOnFailure).
- `mochaSetup.mjs` — global setup: headless Chrome WebDriver and runner.
- `runner.mjs` — starts/stops the WackyTracky server (build frontend, build service, spawn server, wait-on port 8443).
- `tests/` — Mocha test files (e.g. `tests/pageTitle/pageTitle.mjs`).

## Single test

```bash
cd integration-tests && npx mocha tests/pageTitle/pageTitle.mjs -t 15000
```
