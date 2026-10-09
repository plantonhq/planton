---
title: Reading a Run — Status, Stages, Gates, and the Honest Skips
description: How to read one run's record and report it in the user's words — the two run shapes (build runs vs delivery runs), the status vocabulary (queued/running/completed × succeeded/failed/cancelled/skipped and awaiting approval), the build task narrative and its per-task errors, the build logs and how to relay them, the per-environment deploy DAG with the deployment-record join, gate state and who can resolve it, the deploy-stage skip explanations that must be relayed verbatim, the run feed that renders the repository's own CI runs (mirrored from the provider, spoken in the provider's verbatim words) beside Planton's in one chronology, the external CI job logs fetched from the provider on view (completed jobs only), and the external-run verbs (re-run all/failed/one-job with optional debug logging, cancel) that delegate to the provider and resolve on its acknowledgement. Read when someone asks what a run is doing, why it failed, why it did not deploy, whether the repo's own CI passed, what its CI job printed, how to re-run or stop the repo's CI, what is waiting on them, or when you are following a run and reporting progress.
---

For current decisions the caller can make, or personal Read/Unread state,
use `references/craft.organization-activity.md`. A failed run is not an
approval, and an environment's approver roster is not an eligibility check.

# Reading a Run — Status, Stages, Gates, and the Honest Skips

A run (the ServicePipeline record) is one automation run on a service, and its record is written to be READ: real diagnoses in `status_reason` fields, per-task errors on the build graph, per-resource reasons on the deploy graph, and explained skips instead of silent absences. This file is how to read one and answer in the user's words — "your push is building", "staging is waiting on Priya's approval", "it built but didn't deploy, and here is why".

Get one run with `get_service_pipeline` (the FULL record, including the compiled pipeline YAML and the captured per-environment manifests). For "what ran for this service?", prefer `list_service_runs` (CLI: `planton service runs <service>`) — the ONE chronology across run origins: Planton-managed runs AND the repository's own CI runs, newest first (see the external-runs section below). `list_service_pipelines` (CLI: `planton service pipelines`) lists exactly the Planton-managed run records, as projections. The record answers most questions without any other read.

## The two run shapes

Read `spec.trigger.type` FIRST — it decides which stages exist:

- **Build runs** — `webhook`, `manual`, `rerun` — build first (`status.build_stage`), then deploy what they produced. Their story leads with the commit: `spec.git_commit` carries the message, branch, sha, author, and — for pull requests — the PR number and URL.
- **Delivery runs** — `promote`, `rollback`, `cli_deploy` — deploy an artifact that ALREADY exists and have **no build stage at all**. The artifact rides `spec.trigger.artifact` (stamped at birth by the door that created the run), with its source commit as provenance; `spec.trigger.source_service_deployment_id` names the deployment it promoted from or rolled back to. Never report a delivery run as "build missing" — nothing is missing; that is its shape.

`spec.deploy_environment`, when set, scopes the run to exactly ONE environment — no promotion walk. A pull-request run with it set is deploying into its preview environment (the slug is `{service}-pr-{number}`).

## The status vocabulary

Every level — run, stage, environment, build task, deploy resource — speaks the same two enums: a status (`queued`, `running`, `completed`, `awaiting_approval`) and, once completed, a result (`succeeded`, `failed`, `cancelled`, `skipped`). Report them plainly: queued means waiting for a runner; `awaiting_approval` means a HUMAN decision is pending (see gates); `cancelled` carries who and when in the run's `status_reason`.

The `status_reason` at each level is the engine's real diagnosis, written for the user. Prefer the deepest one that exists — a failed run's stage reason names the failed environment; a failed task's `execution.error` on the build graph is the actual error text. Relay these; never summarize a diagnosis into vagueness.

## Reading the build stage

`status.build_stage.dag` is the task graph in execution order (`topological_order` is authoritative). Each task node's `execution` carries its status, result, timings, and — on failure — `error`, the failing step's real output. The build's product is `build_stage.artifact`: the exact image reference the deploys reference, with its source commit.

## Reading the build logs

When the task's `error` summary is not enough — "why did yarn build fail", "what did the migration print" — read the actual build output with `get_service_pipeline_logs`. One call serves both eras: a finished build returns its complete log (the logs outlive the run's infrastructure), and an in-flight build returns everything so far — call again for updates. Lines are attributed per task (the same task names the build graph carries), arrive in true output order, and are RAW: relay the relevant lines verbatim, exactly as the tool returned them — the build's own words are the diagnosis, and a paraphrase drops the detail the developer needs. The console renders these same logs on the run page, where each build step expands into its log and every line has a shareable link — when pointing a human at a specific line, the run page's line link is the better handoff than a quoted wall of text. Scope honesty: these are BUILD logs only; a deploy node's logs live with its Infra Job (read the node's `infra_job_id` from the run and use the infra-job tools). A run with no retrievable logs answers plainly that none are available — report that as the fact it is, never as an error.

## Reading the deploy stage

`status.deployment_stage.environments` renders the promotion walk in order. Each environment carries its own status/result/reason, its resource DAG (each node an Infra Component with per-resource status, its Infra Job id, and outputs), and — the join that answers "so what is running now" — `service_deployment_id`: the deployment record this environment's deploy produced, where the URLs and the rollout verdict live. An environment may also carry its serving hostname; trust `serving_hostname_carried` before treating it as a working address.

## The skips are explanations, not absences

A deploy stage (or one environment) with result `skipped` ALWAYS carries the reason in its `status_reason`, authored in the user's language: pull-request runs without preview deploys enabled, tag builds without tag deploys enabled, a branch outside the trigger set, deployments disabled on the service, no deploy configuration, no artifact produced. **Relay the explanation verbatim** — each names its own remedy, the classes grow over time, and a paraphrase that drops the remedy turns an explained skip back into a mystery.

## Gates: what is waiting, and on whom

A protected environment pauses its run with status `awaiting_approval` and `requires_manual_gate: true`; individual resources can pause the same way on node gates declared in the deploy graph. `list_service_pipelines_awaiting_my_approval`-shaped reads (the queue is organization-scoped) answer "what is waiting on ME" with the exact predicate the resolve guard enforces.

Resolution is a human act: `resolve_service_pipeline_env_gate` / `resolve_service_pipeline_node_gate` (CLI: `planton service resolve-env-manual-gate` / `resolve-node-manual-gate <run-id> <env> [node-id] yes|no`, node ids as `{KindName}/{slug}`). Resolution is idempotent — re-resolving reads back as already resolved, never an error — and separation of duties is the server's law: the person who set the run in motion cannot approve it into a protected environment, and the ASSISTANT can never approve anything, anywhere. When a gate resolves, the environment records the decision, the resolver, and their note — the durable "approved by" answer.

## Following a live run

`run`/`rerun`/`cancel` act; reading reports. A rerun re-executes the source run's stamped pipeline definition byte-identically and answers with the NEW run's id — follow that one. The status stream (the CLI's `follow`, the console's live page) carries status only; anything from the spec — the compiled definition, the captured manifests — comes from the one full get. The compiled definition (`spec.resolved_pipeline`: the exact YAML, its source, and its pin) is the forensic answer to "what actually executed", stamped at dispatch and immune to later changes in the repository or the platform.

## The run's own check on the GitHub commit

A build-lane run of a GitHub-connected service also tells its story on GitHub itself: a check named by the SERVICE's slug appears on the commit the moment the run is born, ticks live (queued, in progress, awaiting approval), and concludes with the run's verdict — the check's output carries the build outcome and every environment's outcome with its reason, and "Details" opens the run in Planton. Three facts shape answers about it: a run parked at an approval gate keeps its check IN PROGRESS with the waiting line (the check never concludes while the run can still change); delivery runs — promote, rollback, `service deploy` — write NO check, deliberately (they deploy an older commit whose check must not be restated; their story is the GitHub Deployment the repo's environments panel shows); and a check missing entirely usually means the GitHub App installation lacks the 'Checks: Read and write' permission — an administrator grants it on the App's settings, then installations re-accept.

## The repository's own CI in the run feed

A service bound to a repository that runs its own CI (GitHub Actions) sees those runs too — mirrored from the provider's webhooks into their own records and rendered in the same feed, with zero setup. `list_service_runs` returns one newest-first page of BOTH: each entry's populated arm names the origin (`planton_run` or `external_run`), so a failed Actions run sits directly above the deploy that followed it, in one chronology.

Reading an external run entry, three rules:

- **Speak the provider's words.** The record carries the provider's own vocabulary verbatim — `provider_status` and `provider_conclusion` are GitHub's words (`in_progress`, `success`, `failure`, `action_required`, `neutral`, ...) and are what you relay; the platform's mapped status exists for machinery, not for people. The spec's provider arm carries the workflow's name and file, the run number, the ATTEMPT (a re-run on GitHub is a new attempt and a NEW record — attempt 1's record stands as history), the trigger event's own name, the commit, the actor, and `html_url` — the deep link to GitHub's own page, always worth handing over.
- **The feed strips the job tree; `get_external_ci_run` carries it.** When someone asks WHICH job or step failed, get the full record: `status.jobs[]` with per-job and per-step status, conclusions, and timings, each job with its own GitHub deep link.
- **A non-empty `reconcile_note` is the record's honesty text.** It means the provider stopped answering for this run (deleted, or past the provider's retention) and the platform closed the record WITHOUT inventing a conclusion — relay the note as the answer, never guess what the run "probably" did.

When the job tree's conclusions are not enough — "what did the test job actually print" — read the job's log with `get_external_ci_run_logs` (CLI: `planton service logs <extrun-id>`, with `--job` to name the job). The platform fetches the log from the provider AT THAT MOMENT (nothing is stored platform-side), so three facts shape the answer: it works for COMPLETED jobs only — GitHub publishes a job's log only after the job completes, so a running job answers with that fact, never a partial log (statuses still tick live from webhooks; ask again when the job finishes); without a job name the single FAILED job is chosen — the diagnosis default — and a run with several candidates answers with their names; and a log the provider no longer serves (retention expiry, deleted run) answers plainly that it is gone — report that as the fact it is. Lines arrive verbatim with the provider's own per-line timestamps: relay the relevant lines exactly, the same rule as build logs. The console renders these same logs on the external run's own page, where each completed job expands into its log.

You can also ACT on an external run, at the provider's own granularities. `rerun_external_ci_run` (CLI: `planton service rerun <extrun-id>`) re-runs a COMPLETED run on GitHub — all jobs by default, only the failed jobs and their dependents with `failed_only` (`--failed-only`), or one named job and its dependents with `job` (`--job`); `enable_debug_logging` (`--debug-logging`) turns on GitHub's diagnostic output for the new attempt, worth offering when someone is chasing a flake. `cancel_external_ci_run` (CLI: `planton service cancel <extrun-id>`) stops an in-flight run — GitHub finishes the step it is on, then stops. Three facts shape how you report these verbs: a re-run's new attempt is a NEW record (the attempt law) that GitHub does not identify synchronously — say "requested; the new attempt appears in the run feed as GitHub reports it" and watch `list_service_runs`, never invent an id; the record after a cancel converges as GitHub reports it, never instantly; and a refusal carries its honest reason — a run still in progress, GitHub's own state refusal relayed verbatim, or the GitHub App installation missing its 'Actions: Read and write' permission (an administrator fixes that on the GitHub App's settings, then installations re-accept) — relay the reason as the answer.

Capability honesty: Planton mirrors these runs, it does not execute them — there are no gates, no compiled definition, and no promotion walk on an external run; status is live from the provider's webhooks, with a background reconciliation healing any missed delivery within minutes. A service whose repository has no external CI simply never sees external entries — the feed is the same read either way. In the console, an external run opens its own GitHub-native page (workflow, run number, attempt, jobs, steps — GitHub's words, GitHub deep links, the same re-run/cancel verbs) through the same run route Planton runs use.
