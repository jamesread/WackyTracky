#!/usr/bin/env python3
"""
Take 3 marketing screenshots of the WackyTracky app using repo-common-take-screenshot.py.

Shots: list-work (Work list), task-options (same list with task details overlay
opened via script), and tpp (/options/task-property-properties). Screenshots are
written to var/marketing/screenshots/. The task-options shot uses
var/marketing/scripts/open-task-options-for-screenshot.py to right-click the
first task and open the task details panel.

Usage:
  # From repo root, server running (config pointing at var/marketing/marketing-todotxt):
  python var/marketing/scripts/take-marketing-screenshots.py

  # Extra --script run before each capture (in addition to the built-in task-options script):
  python var/marketing/scripts/take-marketing-screenshots.py --script path/to/setup.py

  # Custom repo-common path (default: REPO_COMMON env or ../../repo-common):
  REPO_COMMON=/path/to/repo-common python var/marketing/scripts/take-marketing-screenshots.py
"""

import os
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parent.parent.parent.parent
VAR_MARKETING = REPO_ROOT / "var" / "marketing"
SCREENSHOTS_DIR = VAR_MARKETING / "screenshots"
BASE_URL = "http://localhost:8080"
SCRIPT_DIR = Path(__file__).resolve().parent

# List ID from var/marketing/marketing-todotxt (Work)
WORK_LIST_ID = "a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d"


def repo_common_dir():
    d = os.getenv("REPO_COMMON")
    if d:
        return Path(d)
    return REPO_ROOT / ".." / ".." / "repo-common"


def take_screenshot(name: str, url: str, script_paths: list[Path], screenshot_dir: Path) -> None:
    take_script = repo_common_dir() / "repo-common-take-screenshot.py"
    if not take_script.is_file():
        print(f"Error: {take_script} not found. Set REPO_COMMON or run from repo with repo-common at ../../repo-common.", file=sys.stderr)
        sys.exit(1)

    env = os.environ.copy()
    env["SCREENSHOT_DIR"] = str(screenshot_dir)

    cmd = [sys.executable, str(take_script), name, url]
    for p in script_paths:
        if p.is_file():
            cmd.extend(["--script", str(p)])
        else:
            print(f"Warning: script not found {p}", file=sys.stderr)

    result = subprocess.run(cmd, cwd=str(repo_common_dir()), env=env)
    if result.returncode != 0:
        sys.exit(result.returncode)


def main() -> None:
    import argparse
    p = argparse.ArgumentParser(description="Take 3 marketing screenshots of the app.")
    p.add_argument("--script", action="append", dest="scripts", default=[], metavar="PATH",
                   help="Path to Python script defining run(driver); can be repeated. Run before each screenshot.")
    p.add_argument("--url", default=BASE_URL, help=f"Base URL of the app (default: {BASE_URL})")
    args = p.parse_args()

    base = args.url.rstrip("/")
    list_work_url = base + "/lists/" + WORK_LIST_ID
    task_options_script = SCRIPT_DIR / "open-task-options-for-screenshot.py"

    extra_scripts = [Path(s).resolve() for s in args.scripts]
    SCREENSHOTS_DIR.mkdir(parents=True, exist_ok=True)

    tpp_url = base + "/options/task-property-properties"
    shots = [
        ("list-work", list_work_url, []),
        ("task-options", list_work_url, extra_scripts + [task_options_script]),
        ("tpp", tpp_url, []),
    ]

    for name, url, script_paths in shots:
        take_screenshot(name, url, script_paths, SCREENSHOTS_DIR)

    print("Screenshots saved under", SCREENSHOTS_DIR)


if __name__ == "__main__":
    main()
