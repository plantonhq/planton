---
title: "Home Shows What Changed in Your Organization"
date: 2026-09-30
category: feature
tags:
  - console
  - cli
  - platform
excerpt: "Home is now your organization's live activity: who deployed, changed or granted what, what failed, and what waits on you, with the Planton Assistant ready to explain a failure or catch you up."
author:
  - name: Swarup Donepudi
    title: Founder
---

The first page you land on in your organization now answers three questions in seconds: what changed, what broke, and what is waiting on you. Home replaces the old summary dashboard with your organization's live activity. Every card is a change a person, their CI, or the Planton Assistant made, credited to them with their face, and linked to the resource and the run it touched.

## What You See on Home

- **Waiting on You.** The approvals only you can give, and your own latest attempts that failed. It appears only while something needs you, longest-waiting first, with Review one click away.
- **The activity feed.** Deploys, promotions and rollbacks, infrastructure changes, configuration and secret changes, access grants, and the Assistant's own proposals. At rest, each card is two calm lines. Expand one to see its whole story: the commit, the environments it reached, what it created, changed or deleted in your cloud, and the configuration diff.
- **Live, and filtered your way.** New changes arrive while you watch. Filter by area, environment, person, resource or time, and the address bar keeps the view, so you can share it. Press `?` to see the keyboard shortcuts.
- **Only what you can open.** The feed shows exactly the changes to things you already have access to. A teammate limited to staging sees staging's story, and Home tells them so.

## Ask the Planton Assistant From Home

Type a question in the box at the top, or use one of the Assistant's gestures:

- **Explain** on a failed card opens the Assistant already asking what went wrong, whether anything was left half-changed, and how to fix it.
- **Ask** on a waiting approval asks what saying yes would change. The decision stays yours.
- **Catch Me Up** on the "since your last visit" line narrows the feed to what happened while you were away and has the Assistant brief you on it, beside the very cards it summarizes.

## From the Terminal and Your Coding Agent

`planton activity` shows the same feed as a table. Narrow it to one thing with `planton activity Service orders-api`, or use `--env`, `--area`, `--mine`, `--attention` and `--since 7d`. Add `-o json` for scripts.

```bash
planton activity --since 24h --attention
```

The Planton MCP server gains `list_organization_activity`, so the Assistant and any coding agent connected to Planton can read what changed before they diagnose a problem.

## Why This Matters

- **One place to start the day.** No hunting through services and resources to learn what moved overnight.
- **Failures come with context.** Who changed what just before it broke is one card away, and the Assistant starts from it.
- **Trustworthy by design.** Platform housekeeping never appears as a card, a run that ended always says so, and nobody sees a change to something they cannot open.
