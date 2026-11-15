#!/usr/bin/env python3
"""Copies seed/todotxt to a fresh data dir, filling in relative dates.

Usage: render-seed.py <dest-dir>

`{{today+N}}` in any seed file becomes the ISO date N days from today, so due
dates always look upcoming instead of drifting into "overdue".
"""
import datetime
import re
import shutil
import sys
from pathlib import Path

SRC = Path(__file__).resolve().parent / "seed" / "todotxt"
TOKEN = re.compile(r"\{\{today\+(\d+)\}\}")


def render(text, today):
    return TOKEN.sub(lambda m: (today + datetime.timedelta(days=int(m.group(1)))).isoformat(), text)


def main():
    dest = Path(sys.argv[1])
    shutil.copytree(SRC, dest)
    today = datetime.date.today()
    for path in dest.rglob("*"):
        if path.is_file():
            path.write_text(render(path.read_text(), today))


if __name__ == "__main__":
    main()
