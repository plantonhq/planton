---
title: Fixing a Failed Run From Its Page — The Room, the Read Order, and the Consequence
description: How to answer "Help me fix it" on a service run's page — what the room already told you (the run's identity) and what the visible first turn already carries (the failure in the engine's or GitHub's own words), the fix-it read order for a Planton-managed run (the deepest status_reason, the failure class, the log for a task that ran, the files at the run's commit, a clone and reproduction when the log and files do not settle it), the consequence beat that comes before the cause ("is anything running because of this?" — the environments' deployments, the head deployments' source commits), reading the repository's own GitHub Actions run THROUGH Planton (the job tree, the failed job's log fetched from GitHub, the files at head_sha, whether Planton deployed that commit anyway), the shape of the fix proposal (the file, the exact lines, rerun vs new commit, and the pull request the platform opens on the person's yes), and the quiet brief on a run that did not fail — including a run parked at an approval gate, where you explain what saying yes would deploy and never approve. Read when your standing context says `Surface: a service run's page`, when the conversation opens with "My service ...'s run ... failed" or "...'s own GitHub Actions run failed", or whenever someone asks you to fix, explain, or brief a specific run.
---

# Fixing a Failed Run From Its Page — The Room, the Read Order, and the Consequence

A person is looking at one run of their service and asked you to fix it, or to tell them where it stands. The page already shows the failure; your job is what the best engineer they know would do sitting beside them: read the run's own words, then the log, then the code the run actually saw, say what is and is not running because of it, and hand them the exact change. Read this when your standing context says `Surface: a service run's page`, or whenever someone names a run and asks why it failed. How to READ a run is `references/service.reading-a-run.md`; which of the six build-failure classes it is and who fixes each is `references/service.build-failures.md`; how to reach the repository is `references/service.reading-a-repository.md`. This reference composes those into the room's one workflow; it restates none of them.

## What the room already told you

The standing context's `Where the user is right now:` block carries the run's IDENTITY: `Service:` (the slug every tool and verb takes, with `Organization:`), `Run:` (the id — `svcpipe_…` for a Planton-managed run, `extrun_…` for the repository's own GitHub Actions run mirrored by Planton), `Run origin:` in plain words, `Workflow:` for a mirrored run (GitHub's own "CI #482"), `Commit:` (short sha and the message's first line), `Branch:`, and `Environment:` when the run is scoped to one. Never ask which run, which service, or which commit.

What the room deliberately does NOT carry is the failure. A run's status is live, so the failure rides the VISIBLE first turn instead, minted the moment the person clicked: "My service storefront's run svcpipe_… failed. It was building commit 7f3c2e1 ("bump sharp") from main. The build failed at the task build-image: step build-and-push: process exited with code 1: …" — the engine's own words, the most specific the record had. For a mirrored run: "workflow "CI", run #482, on commit 9a1d0f2 … GitHub's conclusion is failure, and the job build-and-test failed." Those words are your starting point, not your conclusion: the record has more.

## The fix-it read order for a Planton-managed run

Read, then speak; the first turn is one answer after these reads, never a question.

1. **The run's own diagnosis** — `get_service_pipeline` (CLI: `planton get service-pipeline <id>`). Read `status.status_reason` first, then the stage's, then the failed environment's or the failed task's: the deepest one is the truth (`reading-a-run`). Classify it (`build-failures`): four of the six classes are answered from the record alone — a compile refusal (the `[code]` lines ARE the cause), a connection or credential before any pod, no runner, never finalized.
2. **The log, for a task that ran and failed** — `get_service_pipeline_logs` (CLI: `planton service logs <id>`), the failing task's lines relayed verbatim. A deploy failure sends you to the node's infra job instead (`infra_job_id` on the node; the infra-job tools). Never quote a wall of it — the run page has every line one click away; quote the decisive lines.
3. **The files the run actually saw** — at `spec.git_commit.sha`, never today's head: `read_repo_files` with the service and `ref` = that sha (CLI: `planton service repo cat <service> <path> --ref <sha>`). The Dockerfile and the lockfile an install failure names; the pipeline and its tasks for a task failure; the kustomize overlay for a `kustomize-build` failure.
4. **Reproduce, when the log and the files do not settle it** — the clone lane at that commit (`reading-a-repository`): on the web your sandbox has `git`, Node, Python, and Go; on the desktop `planton service repo clone <service> --ref <sha>` and the machine's toolchain.

Stop reading when you can name the cause and the fix. Do not read what you do not need: a compile refusal needs no log, a credential failure needs no files.

## The consequence comes before the cause

The person's first fear is not "why" but "what did this do to my running service". Say it first, in one sentence, and it is always answerable from the records:

- **A failed build deployed nothing.** Each environment still runs what it ran before. Name it: `list_service_deployments` for the service (CLI: `planton service deployments <service>`), each environment's head record with its `artifact.source_commit` — "dev is still on b31c9aa from this morning; nothing changed."
- **A failed deploy may have left an environment half-changed.** The run's `status.deployment_stage.environments[]` says which environments completed, which failed at which resource, and which never started (the promotion walk stops). The failed environment's `service_deployment_id` is present only when a deployment record was written; the head deployment tells you what is actually serving.
- **A failed delivery run (promote, rollback, deploy) changed nothing it did not finish.** Say what the target environment runs now.

Then the cause, then the fix.

## A GitHub Actions run, read through Planton

The repository's own CI run appears on the same route, mirrored (`reading-a-run`'s external-runs section owns the record's shape). Planton mirrors it; it has no diagnosis of its own to add, so every word about the failure is GitHub's — the verbatim `provider_conclusion`, the job and step names. The read order:

1. **The job tree** — `get_external_ci_run` (CLI: `planton get external-ci-run <id>`): which job concluded unsuccessfully, which step within it.
2. **The failed job's log, fetched from GitHub at that moment** — `get_external_ci_run_logs` (CLI: `planton service logs <extrun-id> --job <name>`); completed jobs only, the single failed job chosen when you name none. Relay the decisive lines verbatim.
3. **The files at `head_sha`** — the same `read_repo_files` read, the test or workflow file the log names, the code it exercised.
4. **Whether Planton deployed that commit anyway** — the beat no other surface has: `list_service_runs` shows the Planton run that built the same commit, if one did, directly beside the failed Actions run in the feed; `list_service_deployments` shows whether an environment is RUNNING it. "Your tests failed on 9a1d0f2, and dev is running 9a1d0f2 — Planton built and deployed it while the checks were still failing" is the sentence that changes what the person does next. Say it first.

A non-empty `reconcile_note` means GitHub stopped answering for this run and Planton closed the record without inventing a conclusion: relay the note as the answer, never guess what the run "probably" did.

## The shape of the fix

- **The file and the exact lines.** Quote the change as a diff or as the lines to replace, in the file the run actually used, at the commit it used. A fix that names a file without the lines is not a fix.
- **The commit to make, and whether a rerun is the wrong verb.** A code or compile failure needs a NEW commit — the push starts the run; a rerun of it fails identically, and you say so. A credential, runner, or egress failure is fixed outside the commit — `rerun_service_pipeline` (CLI: `planton service rerun <id>`) is right, and you offer it (`build-failures` owns the judgment).
- **You propose; the pull request is how it lands.** Nothing you read or clone is written back to the repository by you. Once the fix is exact, offer it as a pull request: show the whole file as it should read and the pull request's title and description, ask, and on yes let the platform open it on a branch of its own with the receipt — then say whether it builds on its own (`references/service.opening-a-pull-request.md`). Never a commit to the default branch; never a merge you were not asked for. If the person would rather commit themselves, hand them the file and the commit and offer to watch the run.
- **One cause per answer.** A run with several failed resources usually has one root cause and several consequences; name the root and say which failures follow from it.

## The quiet brief, when the run did not fail

The same room opens on every run. On a run that is running, queued, succeeded, cancelled, or parked at a gate the person clicked the quiet invitation and asked for a brief: what the run did or is doing, where it stands, and what — or who — it is waiting on. Read the record, then answer in that order, in one turn.

A run **parked at an approval gate** is the moment this brief matters most. Say what approval would deploy (the artifact and its commit), what the target environment runs today (its head deployment), what changed between them (the commits, from `list_service_runs` or the repository), and whether the environment before it is healthy on this build (its deployment's rollout verdict, `references/service.urls-and-rollout-verification.md`). Then name the decision as the person's: who can approve (the environment's approvers, on the environment record; once decided, the run's gate carries the resolver and their note) and that YOU cannot — the platform refuses you at every gate, whoever started the run. Offer to watch the rollout once they decide. Never offer to approve, resolve, or "handle" the gate.

## Calibration and surfaces

The profile fact sheet beside the room's facts decides the register (`references/craft.personalization.md`): a `vibe-coder` hears "the build failed because a package you added is not in the lockfile the build installs from; run the install once locally and commit the lockfile", with one thing per sentence and the consequence stated first; a `platform-engineer` hears the task, the exit code, the decisive log line, the file:line to change, the commit to make, and whether a rerun would help — in four lines.

The arm is your concern, never the person's. On the web console every read goes through your tools (the sandbox holds no Planton credential in its shell); the clone lane runs in the sandbox. On the desktop the pinned `planton` CLI is signed in — `planton get`, `planton service logs|runs|deployments`, `planton service repo cat|clone`.
