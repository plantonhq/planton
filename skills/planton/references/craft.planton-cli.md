# The planton CLI — Lookups and Pipeline Diagnosis

The `planton` CLI is your window into the user's Planton: what exists, what
deployed, what failed and why. **This reference is complete for your work —
never explore with `planton help` or guess command names; every command your
duties need is written here exactly as it resolves.** Reach for a command's
own `--help` only when a command FROM THIS PAGE fails, and treat that as a
finding to report, not a license to wander. Everything here is read-only
unless marked otherwise. `chart build` has its own contract
(`build-contract.md`); schema grounding has `component-grounding.md`.
Machine-readable output: most snapshot commands take `-o json` (protobuf
field names) or `-o yaml`.
On the platform-tools arm (no CLI — see SKILL.md, "Know your instruments")
every lookup below has a platform tool twin (`list_*`/`get_*` on charts,
projects, pipelines, stack jobs, connections): same questions, same
diagnosis order, tool-for-command.

The infra commands live under one umbrella: `planton infra <noun> <verb>`
(`infra chart`, `infra project`, `infra pipeline`, `infra stack-job`,
`infra cloud-resource`). `planton chart …` is the one sanctioned shorthand
(same commands as `infra chart`); this page writes chart commands in the
short form and everything else canonically.

## Context — whose org and environment am I in?

```
planton context get                      # effective org + environment
planton context set --org <org> -e <env> # change it (a mutation — confirm first)
```

On local desktop instances the context auto-resolves to the seeded org/env —
commands work with zero setup. Most commands also accept `--org` / `-e`
inline for one-off scoping.

## Grounding lookups (session start, before proposing architecture)

```
planton chart list                        # charts in the catalog (org + official)
planton infra project list -o json        # the org's projects (what has been deployed)
planton infra project get <id-or-name>    # one project, full record
planton env list                          # environments in the org
planton search connections                # provider connections (which clouds/clusters)
planton connection authorization list     # which connections which envs may use
planton secret list -o json               # managed secrets ("env" field = scope)
planton variable list -o json             # managed variables ("env" field = scope)
planton catalog search --server           # the catalog MINUS what the org's catalog
                                          # policy disables (offline default shows all;
                                          # see catalog-availability.md)
```

The secret/variable lists ground `$var`/`$secret` references before you write
them (`config-references.md`); `-o json` matters because only the JSON
records carry each entry's `env`, which decides the reference form. Creating
one (`planton secret set` / `planton variable set`) is a mutation — one
confirmation, same as any other.

`planton env list` includes pull-request PREVIEW environments
(`{service}-pr-{n}`, marked as previews in the KIND column). They come and go
with pull requests, refuse direct deletes, and never join promotion order —
never propose architecture that depends on one. Their story lives in
`references/service.preview-environments.md`.

Read `infra project list` results with an eye on `env` — a project in env
`dev` means that chart's resources exist in dev. Caveat: the list rides the search index, so a
project created seconds ago may briefly be missing; `infra project get`
by id/name is the direct read.

## Checking out real files (charts and deployed projects)

What already exists on the platform is checked out, never re-typed:

```
planton chart checkout <slug> --output-dir <dir>            # a published chart, ready to
                                                            # build/customize (org/<slug>
                                                            # scopes to the org's charts)
planton infra project checkout <id-or-name> --output-dir <dir>  # a deployed project's
                                                            # WORKING COPY, binding included
```

Both default the directory to `./<slug>` when the flag is omitted; in your
workspace, always pass `--output-dir <subfolder>` so the checkout lands as
its own top-level subfolder, never at the workspace root. A project checkout
follows `references/infra.deployed-projects.md` from the moment it lands (the
folder carries the hidden binding); re-running it against the same folder
refreshes the managed files from server truth and leaves yours alone.

## The deployment chain — ids you will meet

`InfraChart → InfraProject (infproj_…) → InfraPipeline (infpipe_…) → per-node
CloudResource → stack job (sj_…)`. Diagnosis walks this chain downward
(the model is explained in `deployment-model.md`).

## Diagnosing a failed deploy (the four-step workflow)

```
# 1. Which pipeline? (newest first)
planton infra pipeline list <project-id-or-name>
#    or skip the lookup: planton infra project last-pipeline <project-id>

# 2. What happened? One-shot snapshot, per-node results:
planton infra pipeline status <infpipe_id>
#    → table: NODE | KIND | STATUS | RESULT | STACK_JOB | REASON
#    -o json for the full record (every node's execution, stack-job ids)

# 3. Why did the node fail? Engine logs, one hop:
planton infra pipeline logs <infpipe_id>              # auto-targets the failed node
planton infra pipeline logs <infpipe_id> --node <slug> # a specific node (slug from status)

# 4. Deeper: the stack job record carries the full error detail:
planton infra stack-job list <cloud-resource-id> -o json
planton get stack-job <sj_id> -o json
planton infra stack-job stream-progress-events <sj_id>   # tail the engine output directly
```

Typical reading of `status` output: a node with result `failed` and a
`REASON` mentioning "No provider connection available" or
"kubernetes-provider-connection … not found" is the wiring class — fix per
`kubernetes-on-cluster.md` / `issue-catalog.md`. A failure naming cloud
concepts (subnets, IAM, quotas) is an infrastructure class — the engine logs
from step 3 carry the provider's own error, and `deployment-model.md`
explains how to map it back to the module and spec field. A failure saying a
resource **already exists** is the orphaned-resource class (an earlier run
created it, then died before recording it in state) — repair by importing,
never by delete-and-retry: read `state-import.md`.

## Large records are files, not pipes

Pipeline, stack-job, and full-resource records run long. Fetch a big record
ONCE into a hidden scratch file inside the folder you were given, then read
and search the FILE with your file tools:

```
mkdir -p .scratch
planton infra pipeline status <infpipe_id> -o yaml > .scratch/pipeline.yaml
planton get stack-job <sj_id> -o yaml > .scratch/stack-job.yaml
```

Never pipe a record through improvised scripts to slice it, and never
re-fetch the same record to look at a different part — the file already has
every part. `.scratch/` is a hidden path, so it stays out of the user's
canvas and file tree; it lives inside your folder, so it is inside your
filesystem boundary. Clean it up when the diagnosis ends if the user asked
for tidy folders; otherwise it is harmless working memory.

## Watching a running deploy (humans; agents prefer snapshots)

```
planton follow <infpipe_id|sj_id> --plain   # auto-detects by id prefix
planton infra pipeline stream-status <id>   # status lines until terminal
```

Streams block until the pipeline finishes — in a session, prefer polling
`status` between other work over holding a stream open.

## Generic escape hatches

Any resource, any kind: `planton get <kind> <id> -o json` (e.g.
`get infra-pipeline`, `get cloud-resource`, `get stack-job`) and
`planton search by-resource-kind <kind>` for kinds without a noun-scoped
list. The noun-scoped verbs above are the preferred, discoverable path.
