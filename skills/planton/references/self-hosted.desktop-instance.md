---
title: The Desktop Instance — Starting, Stopping, Restarting and Diagnosing Planton on a Laptop
description: How a local Planton instance runs on Planton Desktop, the five commands that operate it (planton local start|stop|restart, planton daemon status|logs), what heals by itself and what does not, and the failure signatures with their fix. Read when `planton instance current` names 127.0.0.1, a command against the local instance fails or hangs, the person says the desktop app or their local Planton is down, degraded, or stuck, or a step would otherwise ask the person to quit and reopen the app.
---

# The Desktop Instance

Read this when the Planton the person works with runs on their own machine: Planton Desktop's local instance (`planton instance current` shows `127.0.0.1`, `deployment_kind: local`). You operate it yourself. Never ask the person to quit and reopen the app, restart it, or run a command you can run: every lifecycle action has a command, and each one waits until the instance answers before it returns.

## What runs, and who owns it

The app owns the instance. When the person opens Planton Desktop, it starts a supervisor (`plantond`), and the supervisor runs the parts: `postgres` (or the person's own PostgreSQL server), `temporal`, `redis`, `stigmer-server`, `control-plane` and `runner`. Quitting the app stops all of it, and no data is lost. The commands below never run the supervisor themselves. They ask the app, so the app always knows the state, and the person's window shows what you did.

## The five commands

| Command | What it does | Returns when |
|---|---|---|
| `planton daemon status` | The phase (`starting`, `healthy`, `degraded`, `stopping`), and each part's readiness, its restarts within the last hour, and its message | at once |
| `planton daemon logs [source]` | The end of one part's log from the latest start (`--follow`, `--launch 2` for the start before, `--tail` bytes). With no source, it lists the logs there are. Sources are the parts plus `plantond` and `build-cluster`. It reads files on disk, so it works while the instance is down | at once |
| `planton local start` | Opens the app in the background (or hands off to the open one), and the app starts the instance | the instance is healthy (`--wait 0` returns at once) |
| `planton local restart` | The instance stops and the app brings it straight back: every part starts fresh. No data is touched. With the app closed, this is `start` | the instance is healthy again |
| `planton local stop` | The instance stops and stays down; the app shows "stopped from a terminal" with Start | at once |

`planton local teardown` is not a lifecycle step: it deletes the instance's data for a fresh first run. Use it only when the person asks for a fresh start.

## What heals by itself

- **A part that dies comes back.** The supervisor restarts it within seconds, up to three times an hour. Past that it leaves the part down, and `daemon status` says so. Read that part's log before restarting anything: a part that keeps dying has a cause a restart will not fix.
- **A laptop that sleeps or changes networks.** Database connections are bounded and kept alive, so the instance reconnects on its own once the network is back. Give it a minute after the machine wakes before you act.
- **The supervisor itself.** If it dies, the open app starts a fresh one (three times an hour). If someone stopped it on purpose, the app leaves it down.

## Failure signatures and the fix

| What you see | What it means | Do |
|---|---|---|
| Any command: "the lifecycle daemon … is not running" | The instance is stopped, or the app is closed | `planton local start` |
| `daemon status`: `control-plane` not ready, "the control plane cannot get a database connection, so requests are failing" | Requests fail because the database is unreachable. The `postgres` row says whether the database itself answers | If `postgres` is down, fix the database (for the person's own server, check its host and network). If `postgres` is ready and the line stays for a few minutes, run `planton local restart` |
| `daemon status`: a part "exited again after 3 restarts within the hour, so it stays down" | That part keeps dying | `planton daemon logs <part>`, name the cause to the person, then `planton local restart` once it is fixed |
| An apply, destroy, deploy or rerun: "This instance's workflow engine isn't running, so nothing was started" | `temporal` is down: past its restart budget, or hung. Nothing was written, so there is nothing to clean up | `planton daemon status`; if `temporal` is not ready, `planton daemon logs temporal`, then `planton local restart`, then run the command again |
| `degraded` for more than a few minutes with no restart count rising | A part is up but not ready | `planton daemon logs <part>`, then `planton local restart` |
| `planton local start` says Planton Desktop is not installed | The app runs the instance, and it is not here | Ask the person to install it. On macOS: `brew tap plantonhq/tap`, `brew trust --tap plantonhq/tap`, `brew install --cask planton-desktop` (Homebrew loads a third-party tap only once it is trusted). Anywhere, the download at https://planton.ai/desktop |

## What was verified

On the desktop instance on macOS, live:
- `start` from a closed app with the screen locked;
- `start` into an open app;
- `restart`;
- `stop` staying down;
- a killed part coming back and being counted;
- a killed supervisor being brought back by the app;
- a workflow engine past its restart budget, and a hung one, refused with the sentence above within seconds and nothing written.

The database health sentence and a network drop's recovery were proven by tests, not by cutting a live network. On Windows and Linux, `planton local start` opens the same `planton-desktop://local-instance/start` link the app's installer registers. Confirm the link opens the app on the person's machine before you rely on it there.
