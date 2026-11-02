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
planton chart list                        # charts in the catalog (org + platform), each with its version
planton infra stack list -o json        # the org's Infra Stacks (what has been deployed)
planton infra stack get <id-or-name> --env <env>  # one stack, full record (its declaration and live revision);
                                                 # a name needs --env, an id does not
planton infra stack history <name> --env <env>  # its revisions: every plan applied, newest first
planton infra stack plan <name> --env <env>     # what applying would do now (writes nothing)
planton infra stack review <name> --env <env>   # the request waiting in a protected environment (writes nothing)
planton infra stack status <name>                # its environments path by path, stage by stage: what each
                                                 # runs, its chart version, how far behind or ahead, what is
                                                 # moving in
planton infra stack diff <name> --env <env> --against <env>   # two environments side by side, values first
                                                 # (or --revision N --against-revision M); exit 1 = they differ
planton env list                          # environments in the org
planton promotion-path list               # the order they move in: each Promotion Path's stages (siblings
                                          # in braces), where Promotion Only starts, those on no path
planton service status <service>          # a service's environments path by path, stage by stage: the build each
                                          # runs, rollout, distance ("3 commits behind dev"), runs in
                                          # flight by number; -o json|yaml for the path itself
planton search connections                # provider connections (which clouds/clusters)
planton search infra-components --all-envs -o json  # deployed infra components in every environment;
                                          # an answer with only indexFreshness means none
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
with pull requests, refuse direct deletes, and never sit on a Promotion Path —
never propose architecture that depends on one. Their story lives in
`references/service.preview-environments.md`.

Read `infra stack list` results with an eye on `env` — a stack in env
`dev` means that chart's Infra Components exist in dev. Caveat: the list rides the search index, so a
stack created seconds ago may briefly be missing; `infra stack get`
by id/name is the direct read.

## Changing a deployed stack (mutations)

Every change to a stack is a plan you show first, then apply on the user's
yes for THAT plan (or under their word to carry on without asking, then say
exactly what was destroyed or replaced); none of these runs uninvited:

```
planton infra stack apply <name> [folder] --env <env> -m "<why>" --plan-digest <d>
                              # applies exactly the plan shown and follows its run;
                              # the digest or --yes consents to its destroys and replaces;
                              # --keep / --adopt Kind/name name choices
planton infra stack reconcile <name> --env <env>      # plans putting back what drifted; applies
                              # like apply (-m, --plan-digest; --dry-run stops at the plan)
planton infra stack withdraw <name> --env <env>       # takes a waiting request back
planton infra stack promote <name> --from <env> [--to <env>]   # left out, --to is the next stage
planton infra stack rollback <name> --env <env> [--to-revision N]
planton infra stack upgrade <name> --env <env>
```

A request waiting in a protected environment is decided by a PERSON who
read its review: you carry out the decision they stated, or approve under
their word to carry on without asking -- never a run your own person
started (the platform refuses it). Approving approves every Destroy and
Replace the review shows; at a terminal it asks, and `--yes` alone approves
the current review:

```
planton infra stack approve <name> --env <env> --review-digest <d>
planton infra stack reject <name> --env <env> --review-digest <d> --reason "<their words>"
```

`references/infra.deployed-stacks.md` carries the plan rows, consent,
requests, drift and Reconcile; `references/infra.stack-promotion.md`
carries promote, roll back, upgrade and the CI flow.

## Delivering a service (mutations, except the reads marked)

```
planton service status <service>                        # read: each environment, path by path and stage by stage
planton service promote <service> --from <env> [--to <env>] --dry-run   # what the run would do
planton service promote <service> --from <env> [--to <env>]   # or --deployment <svcdepl_id> --to <env>;
                                                        # left out, --to is the next stage on the path
planton service rollback <service> --env <env> [--to-deployment <svcdepl_id>]
planton service redeploy <service> --env <env>          # what runs, applied again (puts drift back)
planton service deploy <service> --env <env> --image <ref>
planton service drift check <service> --env <env>       # read: compares the cloud, writes nothing
planton service drift show <service> --env <env>        # read: the latest verdicts
```

Each delivery verb takes `--dry-run` (what the run would apply, whether it
waits at a gate, and every removal; nothing saved, no run number taken),
`--reapply-all` and `--follow`; each answers with a run ("Run #129") to
follow. A run removes what the service no longer declares once everything
declared applied. In an unprotected environment a push destroys it unasked;
the other verbs name it and ask at a terminal (`--yes` answers ahead). A
protected environment's gate decides every removal it shows (each is
destroyed unless kept), and a removal that appears after the approval
asks on its own:

```
planton service removal destroy <service> <number> <env> <Kind/name>   # asks; --yes answers ahead
planton service removal keep <service> <number> <env> <Kind/name>      # keeps it on its provider
planton service resolve-env-manual-gate <service> <number> <env> approve --keep <Kind/name>   # the rest is destroyed
```

These decisions are a person's: relay the question and the command, and
answer only on their word (never on a run your own person started), then
say exactly what was destroyed. `references/service.delivery-verbs.md` carries the verbs, the
removals and drift; `references/service.reading-a-run.md` the gates.

## Checking out real files (charts and deployed stacks)

What already exists on the platform is checked out, never re-typed:

```
planton chart checkout <slug> --output-dir <dir>            # a published chart, ready to
                                                            # build/customize (org/<slug>
                                                            # scopes to the org's charts)
planton infra stack checkout <id-or-name> --env <env> --output-dir <dir>
                                                            # a deployed stack's WORKING COPY,
                                                            # binding included
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

# 3b. A passing failure (a timeout, a quota)? Run it again, as it was reviewed:
planton infra pipeline rerun <infpipe_id>              # same revision; asks when it destroys or replaces (--yes)

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
resource **already exists** is the unrecorded-resource class (an earlier run
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

## What would a deploy change? A plan that can never apply

```
planton tofu plan <Kind> <name>              # an OpenTofu component; `terraform plan` likewise
planton pulumi preview <Kind> <name>         # a Pulumi component
planton tofu plan <Kind> <name> --destroy    # what a teardown would remove
```

Each starts a plan-only Infra Job: it reads the live cloud inside the plan,
writes nothing (not the recorded state, not the resource), prints the diff,
and ends. It has no apply step and never waits for approval, so nothing can
make it change infrastructure: run it freely to verify a takeover or show
the person a change before they deploy. A plan answers "what would a
deploy change?"; "was anything changed outside Planton?" is a drift check
(below). It exits 1 when the job fails (so does every command that follows
its job: `apply`, `destroy`, `refresh`), so a burst of plans in a script can
trust the exit code. The plan stays on the job: `planton get infra-job <ij_id>
-o json` carries it under `status.iac_operations.update_preview.snapshot`
(`destroy_preview` for a teardown) -- `summary` counts creates, updates,
deletes and replacements, `resource_diffs` holds each resource's diff.
Platform-tools twin: `plan_infra_component` (`destroy: true` for the
teardown), then `get_infra_job` on the returned id.

Not to be confused with `planton tofu apply --pause-for-approval`: that is a
DEPLOY that plans, then waits for someone to approve before applying -- a
mutation, and its approval is a person's decision, never yours.

## Was anything changed outside Planton? A drift check

```
planton infra component drift check <ic_id>                      # or: <Kind> <name>, as tofu plan takes
planton infra component drift show <ic_id>                       # its latest verdict, without checking again
planton infra stack drift check <stack> --env <env>              # every resource the stack holds
planton infra stack drift show <stack> --env <env>
planton service drift check <service> --env <env>               # every resource the service runs there
planton service drift show <service> --env <env>
```

A drift check compares what Planton recorded when it last applied a
resource with what is really in the cloud, setting by setting, and writes
nothing. It never waits for approval, so run it freely; it is the answer
when the user asks whether the cloud still matches, or suspects someone
edited a resource by hand. A resource check follows its job; a stack check
follows until every resource has its verdict. Each verdict is In Sync,
Drifted (with each change: the setting, the value Planton applied, the
value in the cloud now, a secret only as "Hidden value changed", a
resource deleted by hand as "Deleted Outside Planton"), or Couldn't Check
with its reason. `-o json` and `-o markdown` carry the same. Exit codes:
`0` checked, `1` a resource could not be checked, `2` the check could not
run; `--detailed-exitcode` adds `4` for drift found. A resource Planton
never deployed is refused: there is nothing to compare with.

Drift found means the next plan or deploy that reaches the resource puts
the declaration back, with the reason "Drifted: 3 changes made outside
Planton". Putting it back on purpose is, for a stack's resource,
`planton infra stack reconcile <stack> --env <env>` (a plan you show and
the user approves; read `references/infra.deployed-stacks.md`),
`planton service redeploy <service> --env <env>` for a service's resource
(read `references/service.delivery-verbs.md`), or a fresh deploy for a
resource deployed on its own. Keeping the cloud's change means carrying the
value into the template, the service's configuration or the manifest. Which
one is the user's call. Platform-tools twins: `check_infra_component_drift`,
`get_infra_component_drift`, `check_infra_stack_drift`,
`get_infra_stack_drift`, `check_service_drift`, `get_service_drift`,
`redeploy_service`.

A refresh is different: `planton tofu refresh` writes what it reads into
the recorded state, a platform mutation. Never use it to look for drift.

**A rename is not cosmetic.** `planton rename <Kind> <id> -n <name>` (and
`rename_infra_component`) changes the name the component's IaC module is
given at its next deploy, and most kinds use it as the resource name in the provider:
the next deploy replaces that resource unless the kind's spec has an
explicit cloud-name field holding the current name. Read the kind's schema
for such a field before proposing a rename of anything deployed, and plan
after it.

## What changed, and who changed it

```
planton activity --since 24h -o json            # the organization's feed, newest first
planton activity --env prod --failures --read-state unread -o json # failures not inspected
planton activity decisions -o json            # current decisions; org-wide, omit --env
planton activity --mine --since 7d -o json      # the person's own changes
planton activity <Kind> <name> -o json          # one resource's activity
planton activity --this-session -o json         # everything this agent session changed
planton activity --agent codex --since 7d      # what one coding agent changed (or --agent any)
planton history <Kind> <name>                   # one resource's field-by-field versions
planton history <version-id> -o json           # the exact saved version and introduced diff
planton diff <version-a> <version-b> -o json    # compare two saved versions
planton activity mark-read <activity-id>       # MUTATION: mark its current failure inspected
planton activity mark-unread <activity-id>     # MUTATION: mark its current failure uninspected
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
configuration change's diff with `history <version-id>`. For briefing and
review semantics, read `references/craft.organization-activity.md`: decisions,
unread failures and changes use different scopes. Read-state writes require
explicit intent; the CLI fetches a current observation, not an earlier
reviewed snapshot. Listing never marks anything Read.

When you run `planton` from inside a coding agent, every change you make is
recorded as the person **and** you ("Priya Rao and Claude Code"), with your
session. Before you report work as done, run `planton activity
--this-session -o json` and check that what you changed is exactly what you
meant to change; name anything unexpected. `--session <id>` reads another
session's changes (each card carries `spec.actor.agent.session_id`), and
`--agent` takes `claude`, `cursor`, `codex`, `planton-assistant`, `other`, or
`any`; declared aliases such as `claude-code` also work.
`--this-session` refuses in a terminal that runs inside no agent session.
Use the activity reference's capability checks when a command or server
operation is unavailable; an installed skill does not update the CLI or
backend. Authentication and partial-response failures are not empty history.

## Watching a running deploy (humans; agents prefer snapshots)

```
planton follow <run-id> --plain             # any run, its kind read from the id's prefix
planton infra pipeline stream-status <id>   # status lines until terminal
planton service follow <service> <number>   # a service run by its handle ("Run #128"),
                                            # or by its id; logs, rerun, cancel and both
                                            # gate resolvers take the same two forms
```

Streams block until the pipeline finishes — in a session, prefer polling
`status` between other work over holding a stream open. `planton follow`
reads the id's prefix (`ij_`, `infpipe_`, `svcpipe_`; an `extrun_` run is
GitHub's, so it names `planton service logs` instead) and answers by exit
code: `0` the run is done, `1` it failed or was cancelled (or the id names
no run), `3` it waits for a person, and the line before says who runs
what. Branch on the code, never on the words. A command that files a
request instead of applying (an apply, rerun or undeploy into a protected
environment) ends `0` with the review lines: the filed request is its
outcome, and deciding it is a person's act. `3` is a run already in
flight that stopped for a person.

## Running something again: `planton rerun`

```
planton rerun <run-id>                      # any run; follows it to the end
planton rerun <run-id> --no-follow          # starts it and returns
```

`planton follow`, `planton cancel` and `planton rerun` take any run's id
and read its kind from the prefix: `ij_` an Infra Job, `infpipe_` a stack
run, `svcpipe_` a service run, `extrun_` an external CI run. Any other id
exits 1 naming the four. A rerun is a platform mutation: say what it runs
again and confirm first. When the rerun still destroys or replaces
something it asks; `--yes` answers, once the user agreed (or told you to
carry on without asking). `--failed-only`, `--job` and
`--debug-logging` work only for an external CI run and are refused
elsewhere. Exit codes: `0` done, or a request is waiting for its approver
(relay the review link); `1` it failed or was refused; `3` a person must
decide. The noun-scoped forms (`planton infra pipeline rerun`,
`planton service rerun`) behave the same.

## Reading a value in a script

A lookup that finds nothing is an answer with its own exit code, and its
sentence goes to stderr, so stdout carries only a value that exists:

```
v=$(planton variable get db-host -o plain)   # exit 0 found, 3 not found, 1 could not ask
planton secret get db-password --reveal -o plain   # same contract; so do env get and planton get <Kind> <id>
```

Branch on 3 for "not declared" (create it, or fall back) and treat 1 as a
failure to reach or ask the instance -- never parse the banner's words.

Not found also covers what you are not allowed to know about. A record in an organization you are not part of answers exactly like a record that does not exist: the same exit 3 and the same sentence, whether you named it by id or by organization and slug. So exit 3 against another organization never proves a name is free there; check the instance and organization first (`planton context get`). Inside an organization you belong to, a record you may not read is a refusal, not a miss: the Access Needed banner names the roles that could and who can grant them, and the command exits 1.

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

## The company's logo

```
planton branding show                                       # the logo shown, its served addresses, who set it
planton branding set-logo logo.png [--dark-theme light.png] # one change sets both versions
planton branding clear-logo [--dark-theme]                  # back to Planton's mark, or drop only the dark-theme version
```

The scope follows the instance: on a self-hosted install one logo covers the
whole installation (installation administration access; `--org` is refused),
and on planton.ai each organization carries its own (`--org` or the current
context; Owner or Infrastructure Admin; the console's Settings → General →
Logo). `set-logo` and `clear-logo` are mutations: confirm first. When a logo
would be hard to see on the dark theme, `set-logo` and `show` say so and print
the `--dark-theme` command to run. The craft is
`self-hosted.branding.md`.

## Connecting a cloud

```
planton connect gcp --project <project> -o json      # the plan; writes nothing ("written": false)
planton connect gcp --project <project> --print-documents   # the exact setup script and every permission
planton connect gcp verify <slug> -o json            # what a connection can deploy, and the trust it presents
planton connect gcp cleanup <slug> -o json           # what deleting it leaves in the cloud, and the removal script
planton connect aws --profile <profile> -o json      # AWS: the account the profile signs in to
planton connect azure --subscription <id-or-name> -o json   # Azure
planton connect digitalocean --print-documents       # the token's scopes; the token comes from DIGITALOCEAN_TOKEN or stdin
planton connect cloudflare --print-documents         # the same for CLOUDFLARE_API_TOKEN
planton connect gcp detect                           # deploy with this machine's sign-in instead (Runner Identity)
planton delete GcpProviderConnection <slug>          # delete (AwsProviderConnection, AzureProviderConnection, DigitalOceanProviderConnection, CloudflareProviderConnection)
```

The bare `connect` verbs and `delete` are mutations; `verify`, `cleanup` and every `-o json` without `--yes` are read-only. `aws` and `azure` take the same `--print-documents`, `verify`, `cleanup` and `-o json` as `gcp`. The whole journey -- what to explain, the plan, the yes, the verdict, every refusal -- is `infra.connecting-clouds.md`.

## The organization's custom domain

```
planton custom-domain get                       # the organization's domain and its state (exit 3: none)
planton custom-domain set <name> --no-wait      # set it; prints the one record to add
planton custom-domain prompt -o json            # the setup guide: the record, the DNS host, its command
planton custom-domain verify --wait --timeout 8m -o json  # Check Now until Live (exit 0) or a diagnosis (exit 2)
planton custom-domain delete --force            # remove it
```

Team plan, hosted Planton only; `--org` names the organization. `set` and
`delete` are mutations: confirm first. The whole loop, every state's fix and
the DNS change are `organization.custom-domain.md`.

## Generic escape hatches

Any resource, any kind: `planton get <kind> <id> -o json` (e.g.
`get infra-pipeline`, `get infra-component`, `get infra-job`) and
`planton search [text] --kind <kind>` (a platform kind: `service`,
`variable`, `infra_component`, …) for kinds without a noun-scoped list. The
noun-scoped verbs above are the preferred, discoverable path.
