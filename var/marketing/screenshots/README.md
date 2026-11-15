# Marketing screenshots

The PNGs here are the screenshots used in the project README and docs. They
are captured from a throwaway WackyTracky server seeded with fictional data,
so every run produces the same images (apart from due dates, which are always
a few days ahead of the day you run it).

## Regenerate

From the repo root:

```bash
make screenshots
```

or `make` in this directory. This builds the frontend and service, starts a
server on `127.0.0.1:18480` from `tmp/`, captures every image listed in
`screenshots.ini`, writes `screenshot-manifest.yaml`, copies the images into
the Antora docs, and stops the server again, even if a step fails.

Requirements: Go, Node.js, `repo-helper`
and a Chrome/Chromium with a matching `chromedriver`.

To look around the seeded app in your own browser, run `make build seed up`,
open <http://127.0.0.1:18480/> (login `admin` / `admin`), then `make down`.

## How it works

| File | Purpose |
|------|---------|
| `seed/todotxt/` | Fictional todo.txt data with fixed list and task IDs, so URLs are stable. `{{today+N}}` becomes a date N days from today. |
| `render-seed.py` | Copies the seed into a fresh data dir and fills in the dates. |
| `service.yaml` | Config for the throwaway server. Its IAM database (and the default `admin`/`admin` login) only exists under `tmp/`. |
| `screenshots.ini` | One section per image for `repo-helper screenshot`, plus the `title` and `description` used in the manifest and README. |
| `scripts/` | Selenium steps run before a capture: sign in, pin the theme, wait for data, open the task dialog. |
| `generate-manifest.py` | Writes `screenshot-manifest.yaml`. |
| `screenshot-manifest.yaml` | Generated index of the screenshots for indexers. Do not edit it by hand. |

The server runs with `HOME` set to `tmp/run/home`, so its git sync SSH key
setup never touches your `~/.ssh` or global git config.

To add or change a screenshot, edit `screenshots.ini` (and the seed data if
needed), run `make`, and update the README captions to match the descriptions.
