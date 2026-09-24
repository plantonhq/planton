# GcpCloudRunJob Guide

The judgment this guide protects: this resource owns the job DEFINITION,
not the runs. Executions are separate API objects created by triggers —
`gcloud run jobs execute`, Cloud Scheduler, Eventarc, or the execution
tokens at deploy time — and almost every design mistake here comes from
treating the definition like a run ("apply it again to run it again")
or a run like a definition (encoding one batch's parameters into the
template).

## Where a task's secrets live

Each `env` entry takes exactly one of `value`, `valueFromSecret`, or
`secretValue`. `value` is written into the task template, readable by
anyone who can view the job -- configuration only, and on Planton a
`$secret/...` reference there is refused before anything deploys (the
platform resolves references to plain values before the module runs, so it
would land in the template as the secret itself). `valueFromSecret` reads a
Secret Manager secret you already own; `secretValue` (on Planton, only a
`$secret/...` reference) has the component keep the value in a secret of
its own, replicated in the job's region, readable
only by the task identity, and pinned to the version stored -- so a new
value is a new template, and the next execution is the first to see it.
Use `secretValue` for anything referenced as `$secret/...`, and give the
job its own `template.serviceAccount`: without one the grant lands on the
project's Compute Engine default account, shared by every default-identity
workload in the project.

## Idempotence is the task's job, not the platform's

Cloud Run retries each task up to `maxRetries` times (default 3), and a
retry re-runs the container from the start with a fresh timeout budget.
Every task must therefore be safe to run twice: write outputs keyed by
`$CLOUD_RUN_TASK_INDEX`, upsert instead of insert, or stage-then-commit.
`maxRetries: 0` is the honest setting for genuinely non-idempotent work
— fail fast and page a human rather than double-charge a customer.
`parallelism` is the concurrency brake for tasks that hammer one
downstream (a database, an API quota); leaving it unset means "as many
as possible", which is usually what melts the database.

## Execution tokens: deploy-and-run vs deploy-and-verify

The tokens make a deploy itself trigger a run, declaratively.
`startExecutionToken` counts the job READY when the execution STARTS —
fire-and-forget kickoff. `runExecutionToken` counts it READY only when
the execution COMPLETES — which turns the deploy pipeline into the
verifier: a migration job with a run token fails the deploy when the
migration fails, exactly where you want the red X. The token is a
suffix: change it on a later update to trigger another run; an unchanged
token triggers nothing. Set at most one (the spec enforces it), and keep
job name + token under 63 characters combined.

## The startup probe earns its place on slow-starting workers

Jobs have exactly one probe type: startup (HTTP, TCP, or gRPC against
the container's declared port), whose whole window is capped at 240
seconds. Its value is sidecar coordination and slow initializers: a
worker whose sidecar proxy must be up first (`dependsOn` waits on the
sidecar's startup probe), or a model-loading container that should be
killed-and-retried when initialization hangs rather than burn the whole
`timeoutSeconds` budget doing nothing. A worker that simply starts and
runs needs no probe at all — exit code is the health check.

## Timeout arithmetic nobody does until it bites

`timeoutSeconds` (default 600) is per ATTEMPT, not per task: a task may
consume `(maxRetries + 1) × timeoutSeconds` of wall clock before it
finally fails. An execution's worst case is that, times the number of
task waves (`taskCount / parallelism`). Size the timeout from the p99 of
one attempt, not from the batch window — and remember each retry starts
the clock over, so a 6-hour timeout with 3 retries is a full-day worst
case for ONE task.

## Destroy stance

`deletionProtection` (default true) is the API-side guard.
`deletionPolicy` layers the engine-side stance under it: PREVENT fails
destroying plans outright; ABANDON drops the job from state and leaves
the definition (and its execution history) in GCP — break-glass for
state surgery, not an operating mode. Deleting a job deletes its
execution history with it; if the audit trail matters, export logs
before the destroy, because Cloud Logging retention is then the only
copy.

## On the diagram

The job consumes `GcpProject`, `GcpServiceAccount` (task identity),
`GcpKmsKey` (image CMEK), `GcpCloudSql` (socket volumes),
`GcpGcsBucket` (FUSE volumes), and `GcpVpcNetwork`/`GcpSubnetwork`
(direct VPC egress). Nothing references the job downstream today — it is
a leaf that triggers own it: Cloud Scheduler for cron, Eventarc for
event-driven runs, pipelines through the execution tokens.

## Pairs well with

- `GcpCloudRun` — the request-serving sibling; a service that needs
  periodic heavy work pairs with a job instead of running it inline.
- `GcpServiceAccount` — a dedicated task identity scoped to exactly the
  data the batch touches.
- `GcpCloudSql` — managed Unix-socket volumes for ETL against Cloud SQL.
- `GcpGcsBucket` — input/output staging through FUSE volumes with
  `subPath` per batch.
- `GcpCloudSchedulerJob` — cron-triggered executions without a single
  line of trigger code.
