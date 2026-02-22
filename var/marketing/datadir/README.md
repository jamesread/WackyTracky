# Marketing / screenshot test data

This directory contains sample todo.txt data for taking app screenshots and marketing materials.

## Usage

Point the backend at this directory via your config:

- **Config:** Set `database.driver` to `todotxt` and `database.database` to this directory (e.g. absolute path like `/path/to/WackyTracky/var/marketing/marketing-todotxt`, or relative to where you start the server, e.g. `./var/marketing/marketing-todotxt` when run from repo root).
- **Docker:** Mount this directory and set `database.database` to the container path.

Then start the app and open the UI; you’ll see:

- **Lists:** Inbox, Work, Personal, Shopping, Goals
- **Tasks:** A mix of priorities (A/B/C), contexts (@work, @home, @phone, …), tags (#urgent, #someday), due dates, and one parent task (“Read 12 books this year”) with two subtasks
- **Completed:** A few items in the completed/done view

You can edit or extend the files (todo.txt, done.txt, todotxt_lists.txt) as needed for more screenshots.
