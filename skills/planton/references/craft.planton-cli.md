# The planton CLI — Lookups and Pipeline Diagnosis

The `planton` CLI is your window into the user's Planton: what exists, what
deployed, what failed and why. **This reference is complete for your work —
never explore with `planton help` or guess command names; every command your
duties need is written here exactly as it resolves.** Reach for a command's
own `--help` only when a command FROM THIS PAGE fails, and treat that as a
finding to report, not a license to wander. Everything here is read-only
unless marked otherwise. `chart build` has its own contract
(`infra.build-contract.md`); schema grounding has `catalog.kind-grounding.md`.
Machine-readable output: most snapshot commands take `-o json` (protobuf
field names) or `-o yaml`.
On the platform-tools arm (no CLI — see SKILL.md, "Know your instruments")
every lookup below has a platform tool twin (`list_*`/`get_*` on charts,
Infra Stacks, pipelines, Infra Jobs, connections): same questions, same
diagnosis order, tool-for-command.

The infra commands live under one umbrella: `planton infra <noun> <verb>`
(`infra chart`, `infra stack`, `infra pipeline`, `infra job`,
`infra component`). `planton chart …` is the one sanctioned shorthand
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
planton infra stack list -o json        # the org's Infra Stacks (what has been deployed)
planton infra stack get <id-or-name>    # one stack, full record
planton env list                          # environments in the org
planton search connections                # provider connections (which clouds/clusters)
planton connection authorization list     # which connections which envs may use
planton secret list -o json               # managed secrets ("env" field = scope)
planton variable list -o json             # managed variables ("env" field = scope)
planton secret list --env <env> -o json   # what <env> can read: its own + the org's
planton infra state-backend list -o json  # state backends, with each one's ENCRYPTION (get <slug>
                                          # for one, its key and phase; exit 3 = none)
planton infra state-backend rekey <slug>  # change the key (--key-source ...), rotate
                                          # Planton's passphrase, or resume a change
planton infra state-backend verify <slug> # sign in and write/read/delete as a deploy would;
                                          # -f <manifest> verifies before create (exit 1 = unhealthy)
planton infra state-backend create <slug> --provisioner terraform --type s3 \
  --auth-mode connection --s3-connection <aws-connection> --s3-bucket <b> --s3-region <r> \
  --s3-use-lockfile                       # a bucket reached through a (keyless) connection:
                                          # nothing stored; --gcs-connection for Cloud Storage;
                                          # share the connection with every env it serves
planton catalog search --server           # the catalog MINUS what the org's catalog
                                          # policy disables (offline default shows all;
                                          # see catalog-availability.md)
```

The secret/variable lists ground `$var`/`$secret` references before you write
them (`infra.config-references.md`); `-o json` matters because the JSON records
carry each entry's `env`, which decides the reference form. A list is every
record (never cut off); only a typed `--env` narrows it, to that
environment's records plus the organization's. Creating one
(`planton secret set` / `planton variable set`) is a mutation — one
confirmation, same as any other.

`planton env list` includes pull-request PREVIEW environments
(`{service}-pr-{n}`, marked as previews in the KIND column). They come and go
with pull requests, refuse direct deletes, and never join promotion order —
never propose architecture that depends on one. Their story lives in
`references/service.preview-environments.md`.

Read `infra stack list` results with an eye on `env` — a stack in env
`dev` means that chart's Infra Components exist in dev. Caveat: the list rides the search index, so a
stack created seconds ago may briefly be missing; `infra stack get`
by id/name is the direct read.

## Checking out real files (charts and deployed stacks)

What already exists on the platform is checked out, never re-typed:

```
planton chart checkout <slug> --output-dir <dir>            # a published chart, ready to
                                                            # build/customize (org/<slug>
                                                            # scopes to the org's charts)
planton infra stack checkout <id-or-name> --output-dir <dir>  # a deployed stack's
                                                            # WORKING COPY, binding included
```

Both default the directory to `./<slug>` when the flag is omitted; in your
workspace, always pass `--output-dir <subfolder>` so the checkout lands as
its own top-level subfolder, never at the workspace root. A stack checkout
follows `references/infra.deployed-stacks.md` from the moment it lands (the
folder carries the hidden binding); re-running it against the same folder
refreshes the managed files from server truth and leaves yours alone.

## The deployment chain — ids you will meet

`InfraChart → InfraStack (infstk_…) → InfraPipeline (infpipe_…) → per-node
InfraComponent (ic_…) → InfraJob (ij_…)`. Diagnosis walks this chain downward
(the model is explained in `infra.deployment-model.md`).

## Diagnosing a failed deploy (the four-step workflow)

```
# 1. Which pipeline? (newest first)
planton infra pipeline list <stack-id-or-name>
#    or skip the lookup: planton infra stack last-pipeline <stack-id>

# 2. What happened? One-shot snapshot, per-node results:
planton infra pipeline status <infpipe_id>
#    → table: NODE | KIND | STATUS | RESULT | INFRA_JOB | REASON
#    -o json for the full record (every node's execution, Infra Job ids)

# 3. Why did the node fail? Engine logs, one hop:
planton infra pipeline logs <infpipe_id>              # auto-targets the failed node
planton infra pipeline logs <infpipe_id> --node <slug> # a specific node (slug from status)

# 4. Deeper: the Infra Job record carries the full error detail:
planton infra job list <infra-component-id> -o json
planton get infra-job <ij_id> -o json
planton infra job stream-progress-events <ij_id>   # tail the engine output directly
```

Typical reading of `status` output: a `REASON` starting "Nothing ran:
environment <env> may not use the <provider> connection <slug>" is a missing
connection authorization — no Infra Job started and nothing in the cloud
changed (the follow says "Refused before any resource ran"); the sentence
carries the `planton connection auth create …` fix, a mutation to confirm.
A node with result `failed` and a
`REASON` mentioning "No provider connection available" or
"kubernetes-provider-connection … not found" is the wiring class — fix per
`infra.kubernetes-on-cluster.md` / `infra.issue-catalog.md`. A failure naming cloud
concepts (subnets, IAM, quotas) is an infrastructure class — the engine logs
from step 3 carry the provider's own error, and `infra.deployment-model.md`
explains how to map it back to the module and spec field. A failure saying a
resource **already exists** is the orphaned-resource class (an earlier run
created it, then died before recording it in state) — repair by importing,
never by delete-and-retry: read `infra.state-import.md`.

## Large records are files, not pipes

Pipeline, Infra Job, and full-resource records run long. Fetch a big record
ONCE into a hidden scratch file inside the folder you were given, then read
and search the FILE with your file tools:

```
mkdir -p .scratch
planton infra pipeline status <infpipe_id> -o yaml > .scratch/pipeline.yaml
planton get infra-job <ij_id> -o yaml > .scratch/infra-job.yaml
```

Never pipe a record through improvised scripts to slice it, and never
re-fetch the same record to look at a different part — the file already has
every part. `.scratch/` is a hidden path, so it stays out of the user's
canvas and file tree; it lives inside your folder, so it is inside your
filesystem boundary. Clean it up when the diagnosis ends if the user asked
for tidy folders; otherwise it is harmless working memory.

## Reading back and comparing a manifest

```
planton get <Kind> <name> -o yaml            # an Infra Component AS ITS MANIFEST: kind, metadata
                                             # (id included), spec, status.outputs; camelCase keys
planton get <Kind> <name> -o yaml --envelope # the stored InfraComponent wrapper instead
planton service get <slug> [-o json]         # the service's stored record (the service.yaml shape)
planton diff -f <manifest>                   # the local file vs what is stored for it
planton validate -f <manifest>               # offline check; every document in the file
```

`get -o yaml` uses the manifest dialect (camelCase: `status.outputs.vpcId`);
`-o json` uses proto field names (`status.outputs.vpc_id`). Read a key by
the name the chosen format prints. A secret output reads as its `$secret/`
reference, never the value.

`diff -f` takes one manifest per file, leaves status and platform stamps out,
and answers by exit code: 0 applying would change nothing, 1 they differ (a
unified diff; `-o json` gives `differs` and `changes[]`), 3 "Nothing Stored
Yet". It is read-only — the check before proposing an apply.

Two refusal banners, two owners: **Manifest Isn't Valid** is the CLI's own
offline check (`apply`, `service register`, `validate`) refusing a file
before sending it — fix the field it names (`validate` on a multi-document
file names the document: "document 2 of 2 (name): …"); **Request Refused** is the server
refusing a request as invalid, nothing changed — relay its reason.

## What changed, and who changed it

```
planton activity --since 24h -o json            # the organization's feed, newest first
planton activity --env prod --attention -o json # what failed or waits for approval in prod
planton activity --mine --since 7d -o json      # the person's own changes
planton activity <Kind> <name> -o json          # one resource's activity
planton activity --this-session -o json         # everything this agent session changed
planton activity --agent claude-code --since 7d # what one coding agent changed (or --agent any)
planton history <Kind> <name>                   # one resource's field-by-field versions
```

`activity` is the answer to "what changed", "what broke" and "who touched
it": one card per change a person, their CI, or the Assistant made, by hand
or through a coding agent, with who, what, when, and how its run ended.
Platform housekeeping never appears,
and every card is trimmed to what the signed-in person may open, so an empty
page is "nothing you may see in this view", never a refusal. It covers the
whole organization; `--env` narrows only when passed. Windows (`--since`,
`--until`) take `24h`, `7d`, `30d`, any `<n>h` or `<n>d`, or an RFC 3339
moment; `--area` takes `infrastructure`, `services_pipelines`,
`connections_credentials`, `configuration_secrets` or
`organization_members`. Each card's `spec.source` names the run behind it:
read a failed service run's logs or an Infra Job from there, and a
configuration change's diff with `history <version-id>`. Lead a summary with
the `--attention` cards, then the rest by area, naming people and resources.

When you run `planton` from inside a coding agent, every change you make is
recorded as the person **and** you ("Priya Rao and Claude Code"), with your
session. Before you report work as done, run `planton activity
--this-session -o json` and check that what you changed is exactly what you
meant to change; name anything unexpected. `--session <id>` reads another
session's changes (each card carries `spec.actor.agent.session_id`), and
`--agent` takes `claude-code`, `cursor`, `assistant`, `other`, or `any`.
`--this-session` refuses in a terminal that runs inside no agent session.
If `planton activity` is an unknown command, the CLI is older than this
feature: tell the person to update it (`brew upgrade planton`) and stop,
rather than piecing the answer together from other commands.

## Watching a running deploy (humans; agents prefer snapshots)

```
planton follow <infpipe_id|ij_id> --plain   # auto-detects by id prefix
planton infra pipeline stream-status <id>   # status lines until terminal
```

Streams block until the pipeline finishes — in a session, prefer polling
`status` between other work over holding a stream open.

## Reading a value in a script

A lookup that finds nothing is an answer with its own exit code, and its
sentence goes to stderr, so stdout carries only a value that exists:

```
v=$(planton variable get db-host -o plain)   # exit 0 found, 3 not found, 1 could not ask
planton secret get db-password --reveal -o plain   # same contract; so do env get and planton get <Kind> <id>
```

Branch on 3 for "not declared" (create it, or fall back) and treat 1 as a
failure to reach or ask the instance -- never parse the banner's words.

## On a self-hosted instance

`planton instance show` reads `deployment_kind: self_hosted`; the same CLI
talks to it, plus the administration verbs. Read freely:

```
planton instance show | list | path          # the manifest, the roster, the credentials directory
planton whoami                               # who the token says you are
planton can-i <action> [--env <env>]         # exit 0 allowed, 2 denied, 1 error; a denial names the role classes
planton access [--for <email>]               # who has access here, or what one principal holds
planton explain-access …                     # why an access holds, path by path
planton directory groups | preview | mappings   # the directory's groups, a rule's blast radius, the rules with sync health
```

Verbs that change access -- `grant`, `invite`, `team`, `service-account`,
`api-key`, `directory map` / `unmap` -- are mutations (confirm, `--yes` only
non-interactively). The craft around them is the `self-hosted.*` references:
front doors and `login --local` (`self-hosted.front-doors-and-the-cli.md`,
`self-hosted.identity-primary-and-break-glass.md`), mapping
(`self-hosted.identity-mapping.md`).

## Generic escape hatches

Any resource, any kind: `planton get <kind> <id> -o json` (e.g.
`get infra-pipeline`, `get infra-component`, `get infra-job`) and
`planton search [text] --kind <kind>` (a platform kind: `service`,
`variable`, `infra_component`, …) for kinds without a noun-scoped list. The
noun-scoped verbs above are the preferred, discoverable path.
