# Repository-Global Locks

This folder holds cooperative lock files that coordinate multiple coding
agents working on the same branch of this repository at the same time. The
folder is committed; the lock files themselves are transient and gitignored
(see `.gitignore` — everything except this README and the `.gitignore` is
ignored).

## What this folder guards (repo-global operations)

| Lock file | Acquire before |
|---|---|
| `git-commit.lock.md` | The session's single wrap-up commit + push |
| `proto-build.lock.md` | Any whole-tree generation or shared-surface edit: `make protos` / chunked `buf generate`, `make generate-catalog-kind-map`, `make reset-gazelle`, `make generate-proto-docs`, `make e2e-matrix`, site catalog regeneration; also brief edits to shared choke-point files — the E2E verifier dispatch switch, the provider E2E test entrypoint files, Makefile test-tier regexes, the catalog-kind enum — and to the shared workflow rules (`_rules/`) |

**Work scoped to your own kind folders NEVER needs a lock.** Locks exist
only for the operations listed above. Two concurrent Bazel builds
(`make build-go`) need no lock either — Bazel serializes on its own workspace
lock.

## How to acquire

1. Check whether the lock file exists. If it exists and is not past its
   `expires` timestamp, wait (do other in-scope work and poll periodically).
2. If absent or expired, write the lock file with this exact shape:

   ```markdown
   owner: <agent/session identifier>
   acquired: <ISO-8601 timestamp>
   expires: <ISO-8601 timestamp>   # git-commit: +10 min; proto-build: +60 min
   purpose: <one sentence — exactly which operation and why>
   ```

3. Re-read the file after writing and confirm the `owner` line is yours
   (guards the simultaneous-create race).
4. Do the guarded operation. **Delete the lock file immediately when done.**
   Locks are held per focused operation, never for a whole session.

## Stale locks

A lock past its `expires` timestamp may be deleted by any agent. Record the
takeover (previous owner, purpose, age) in your own session record so the
orphaned session's owner can reconcile.
