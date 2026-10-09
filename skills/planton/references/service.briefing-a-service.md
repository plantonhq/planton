---
title: Briefing a Service — The Room, the Read Order, and the First Turn
description: How to answer "brief me" on a service the person is looking at — the standing facts a service room gives you and what they save you from asking, the read order (the record, what runs in each environment, the newest runs, the failing run's own words), the shape of a briefing (attention first or one quiet line, what runs where, only the offers you can keep), calibrating it from the profile fact sheet, when to leave the records for the repository, and what each surface can and cannot do when the question needs the files
---

# Briefing a Service — The Room, the Read Order, and the First Turn

A person opens a service's page and asks you to brief them. The page already shows the service; your job is to say, in their register, what they would want a senior engineer to tell them after five minutes with the same console — what needs attention, what is running where, and what you can do next. Read this when your standing context says `Surface: a service's detail page`, or whenever someone asks "how is my service doing" by name.

## What the room already told you

When the conversation was created on a service's page, your standing context carries a `Where the user is right now:` block: the surface, the service's slug (`Service:` — the identifier every tool and verb takes, with `Organization:`), the environments it is declared to deploy into (`Deploys to:`), its repository as `owner/name`, the connection that reaches it (`Repository connection:` — the slug `get_github_access_token` takes when a clone is needed), the default branch, and `Deployments: paused` only when they are switched off. Everything there is a fact from the record the person is looking at. Never ask which service, which organization, or which repository; never look the connection up. Absent lines mean the record has nothing there (a service with no repository binding has no `Repository:` line), not that you should ask.

What the room deliberately does NOT tell you is what is running. That is your first read.

## The read order

Read, then speak; the briefing is one turn after four reads, never a question.

1. **The service** — `get_service` (CLI: `planton get service <slug>`): what it is (description, build, whether deployments are paused, the manifest-sync state when the repository maintains the deploy block).
2. **What runs in each environment** — `list_service_deployments` filtered to the service, taking each environment's newest record (CLI: `planton service urls <service>` answers every environment's current URLs and rollout verdict in one call; `planton service deployments <service>` for the records). Each head record carries the artifact, its source commit, who or what delivered it and when, its URLs, and its rollout verdict — `references/service.urls-and-rollout-verification.md` owns the verdict words; relay them exactly.
3. **The newest runs** — `list_service_runs` (CLI: `planton service runs <service>`): one chronology of Planton runs and the repository's own CI runs. Note anything failed, anything awaiting approval, and whether the newest commit on the default branch has been built yet.
4. **The failing run's own words, when one failed** — the run's deepest `status_reason` (`references/service.reading-a-run.md`); for a build failure, `references/service.build-failures.md` classifies it.

Stop there for the first turn. Do not read logs, files, or the repository before you have briefed — the person asked for a briefing, not an investigation; offer the investigation.

## The shape of a briefing

Service health and personal inspection are different questions. A failed run
can affect this service even after someone reads it, and a historical failure
does not prove the service is currently unhealthy. For "what waits on me" or
unread failures, follow `references/craft.organization-activity.md`: current
decision eligibility comes from the source queues, and Read is personal
inspection. Never label every failed run as waiting for the caller.

- **One line of identity.** What the service is, from where, to where: "storefront is your web storefront from acme/storefront, deployed to dev, staging, and prod."
- **What needs attention, FIRST — or one quiet line.** A failed run, a rollout that did not verify, a run waiting for approval, a configuration sync that failed, a paused switch, a preview that never deployed. Order by consequence to the person, name the fact and its time, and say what is and is not affected ("nothing new was deployed from it"). When nothing needs attention, say so in one sentence and move on — quiet is the signal; never invent concern.
- **What runs where.** Per environment: the build (short commit and its message), how it got there (from a run, promoted, rolled back, deployed by hand), when, its verdict in plain words, and its address. Healthy environments get one line each.
- **The offers — only verbs you can keep.** Read a failure and say exactly what broke; read a run's checks or logs; explain what a waiting build would change; validate the pipeline; promote, roll back, or deploy an image with the person's consent (`references/service.delivery-verbs.md`); read or reproduce from the repository (below). End with one question that lets them choose. One thing you never offer: **approving a gate** — a run waiting for approval is an approver's decision, the platform refuses you at every gate whoever started the run, so name it as waiting for a person and offer to explain what it contains. And one thing you offer only its one way: **a change to the repository** goes out as a pull request the platform opens after the person has seen the exact change and said yes — never a commit of your own (`references/service.opening-a-pull-request.md`).

## Calibration comes from the fact sheet

The profile fact sheet beside the room's facts decides the register (`references/craft.personalization.md`, `references/craft.profile-vocabulary.md`). The same facts, two people:

- A `vibe-coder` with low cloud experience gets plain nouns — "did not pass its health check", "a build waiting for someone to approve it" — one thing per sentence, the consequence stated, and a single confirming question at the end. No `ServicePipeline`, `ServiceDeployment`, `rollout verdict`, or `gate`: if a Planton noun is needed, introduce it with its reason first.
- A `platform-engineer` gets identifiers, timings, and a choice: short commits, environment names, the verdict word, "nothing to roll back to", three next reads named. No explanation of what a thing is.

Their words about the TASK still win: a person who asks about one environment gets that environment's briefing, not the tour.

## When the briefing needs the repository

The records answer the briefing; the repository answers "why". Move to it only after the person picks a thread, and read at the run's commit, never today's head (`references/service.reading-a-repository.md` owns the two lanes and the judgment):

- **One or a few files** — the Dockerfile the failed build used, the pipeline and its tasks, the lockfile a `npm ci` failure names: the platform-mediated read (`read_repo_files` with the service and `ref` = the run's commit; CLI `planton service repo cat`). No credential leaves the platform; works for anyone who can see the connection.
- **Something must run or be searched** — reproduce the install, grep the tree, diff two commits: the clone lane, at the exact commit, with the connection's one-hour token used once and never stored. The room's `Repository connection:` line is the slug to mint against.

Where each runs, by surface — the arm is your concern, never the person's:

- **The web console** (platform-tools arm): your sandbox has `git`, `curl`, Node, Python, and Go, runs as root with `apt` available, reaches the network, and keeps its working directory across turns. Clone and reproduce there. It holds NO Planton credential in its shell: every read of the platform — the service, deployments, runs, logs — goes through your tools, never a CLI. Do not install `planton` there expecting it to sign in; it cannot.
- **The desktop** (CLI arm): the pinned `planton` CLI is signed in to the person's instance and `planton service repo clone <service> --ref <sha>` does the hygienic clone; the machine's own toolchain does the rest.

Nothing you clone or read is written back to the repository by you. When the fix is known, the one way it lands is a pull request the platform opens on your behalf after the person has seen it and said yes — `references/service.opening-a-pull-request.md`.

## When the person opens a run from the briefing

A briefing that names a failed run often ends with the person opening it. The run's own page is a room too: `references/service.fixing-a-failed-run.md` owns what happens there -- the fix-it read order, the consequence before the cause, and the quiet brief on a run parked at a gate.

## When the room is not a service's page

The same read order and shape answer "how is `storefront` doing?" from any surface — you simply start from the name the person gave (`get_service` by organization and slug; CLI `planton get service <slug>`) instead of the room's facts.
