# Deployed Infra Stacks — Working Copies, Plans, Fixing Failures

A chart is a blueprint; an **Infra Stack** is one chart declared into one
environment. A stack is named within its environment, and the same name on
the same Promotion Path is the same stack at another stage: `app-infra` in
qa and in prod on one path are two records — each with its own values,
revisions and runs — and one stack moving through its stages, so every
command that names a stack also names its environment (`--env`). The same
name on another path, or in an environment on no path, is another stack. A stack carries its OWN
declaration — the templates, the chart's `values.yaml` (every value
declared, with its default) and the values set for its environment.
Editing a chart never changes a deployed stack, and editing a stack never
changes the chart. A stack installed or upgraded from a catalog chart
records that chart and its version ("app-infra v3") — the platform's
record, kept beside the declaration and never inside it, carried by every
later edit — so a newer version of the chart reads **Update Available**
and **Upgrade** brings it. Moving a template between environments,
returning to an earlier revision and upgrading are their own reference:
`infra.stack-promotion.md`. This file governs the second mode of your
work: changing something that is already running.

Every change to a stack is a **plan** a person reads first. A plan says
what applying would do to each resource, before anything runs; applying
it runs exactly that plan, and only while nothing it was made from has
moved. Each applied plan is a numbered **revision** ("r12 · Raise
replicas").

## Am I working on a chart or on a deployed stack?

Check for the binding file before assuming:

```
cat .planton/stack.yaml
```

The workspace file tree does not show hidden paths, so LOOK for it with the
shell — never conclude from the tree alone. A folder carrying this file is a
**working copy of a deployed Infra Stack**; the file names the stack
(`stackId`, `stackName`), its organization, and its environment:

```yaml
stackId: infstk_01jb7k4m2x9q3r5t6v8w0yzabc
stackName: eks-mumbai
org: acme
env: prod
```

No binding file → a plain chart folder; this reference does not apply.

A working copy holds:

| File | What it is |
|------|------------|
| `templates/**` | The stack's templates. |
| `values.yaml` | The chart's declarations and defaults, as the stack carries them. |
| `values.<env>.yaml` | Only the values set for the stack's environment (a `params:` list of name and value). Anything not listed takes the default from `values.yaml`. |
| `.planton/stack.yaml` | The binding. |

To change a value for this environment, edit `values.<env>.yaml`; to
change a default every environment shares, edit `values.yaml` (a plan
shows such a change as a value change, marked "the chart's default
changed").

What the binding changes:

- **Your edits target the running stack** — through a plan you show the
  user before anything applies (below). The organization's catalog policy
  never blocks these changes: it refuses only NEW creations of disabled
  kinds, and a deployed stack's Infra Components already exist — never
  warn about availability on a working copy (`catalog.availability.md`).
- **The folder is laid out fresh from the server** — by the app when the
  conversation starts, or by your own checkout — so what you read here is
  the deployed truth as of that moment, placed for THIS conversation alone.
- **The chart is out of the picture.** Never suggest publishing this folder
  as a chart or `chart publish` unless the user explicitly wants to turn
  their stack's declaration into a reusable chart.

## When you are NOT in a working copy: check one out

A conversation about a deployed stack does not always start inside its
working copy — in your own workspace, the request may simply name the
stack. Never reconstruct a stack's files from its record by hand; pull
the real thing:

```
planton infra stack checkout <id-or-name> --env <env> --output-dir <stack-dir>
```

A name alone is ambiguous when the chart is installed in several
environments; pass the environment the user means (an id needs none).

The checkout lays out the working copy as its own top-level subfolder of
your workspace, and everything in this reference applies to it from that
moment. It writes many files at once, so the composing declaration
(SKILL.md's live-screen rule) comes FIRST. Re-running it against the same
folder refreshes the managed files (`Chart.yaml`, `values.yaml`,
`values.<env>.yaml`, `templates/`, `.planton/`) from server truth and
leaves any other files alone.

## Opening posture: diagnose FIRST when there is a failure

When the conversation opens from a failed run (the opening message names a
pipeline id), **the four-step diagnosis is your FIRST ACT — run it before
any reply that merely acknowledges the request.** The user already asked
for the investigation by arriving here; a first turn that restates the
failure and waits is a failed turn. Consent gates CHANGES, never
diagnosis — reading status, logs, records and plans requires nobody's
permission.

1. Run the four-step diagnosis from `craft.planton-cli.md` (status → failed
   node → logs → Infra Job when needed), dumping large records to
   `.scratch/` files and reading those (the large-records rule in
   `craft.planton-cli.md`).
2. Explain what went wrong in the user's language — name the component and
   the cause, not the plumbing.
3. Recommend the fix with a reason, and show its plan (below).
4. **Ask what they want** — even when the fix is obvious. Changing a
   deployed stack is always consent-gated.

The same compile loop applies while editing: `planton chart build . -o json`
works on a working copy exactly as on a chart folder (the binding file is
invisible to it). Keep the folder green before you plan.

## Plan, show, consent, apply

**1. Plan.** Planning writes nothing, so do it freely:

```
planton infra stack plan <stackName> . --env <env> --org <org>
```

Read every flag from the binding file — never guess the name, org, or env,
and never omit `--org`/`--env` (your CLI context may point elsewhere; the
stack's environment is not yours to move). With no folder argument it
plans the declaration the stack already holds — what moved underneath it
(a module a new release changed, a secret's new version, an upstream output). `--set
name=value`, `--set-file name=path` and `-f <file>` change values on top of
the ones the environment already sets.

The plan lists each resource by what applying does to it, in reading
order, each with its reason:

| Row | Meaning |
|-----|---------|
| **Blocked** | It cannot apply as declared; the row says why and the way out. |
| **Destroy** | No longer in the templates: destroyed on its provider once everything the plan applies has applied. |
| **Replace** | Still in the templates, but its preview found a setting its provider cannot change in place: it is destroyed and created again, so what it holds does not carry over. Only a request's review reads it (a preview finds it, below); it needs the requester's destroy rights in the environment, or the row is Blocked. |
| **Update** | Its settings, a secret's version, a value it reads or its module moved — or a drift check found it changed outside Planton ("Drifted: 3 changes made outside Planton"), and applying puts the declaration back. |
| **Update Record** | Only Planton's record changes; nothing in the cloud (an annotation like `planton.dev/prevent-destroy` added). |
| **Create** | New. |
| **Adopt** | It already exists and no stack manages it; this stack takes it over only when chosen, and the apply may change it to match the template. |
| **Keep** (named by its provider: Keep in AWS) | No longer in the templates, and kept: the stack lets go of it, and it keeps running and billing in the account. |
| **May Change** | It reads a value known only after an earlier resource applies. |
| **Always Runs** | Its job runs on every apply: whether it changes is known only when it runs. |
| **Unchanged** / **Unmanaged** | Nothing moved / the platform records it and never applies it. Collapsed into one line. |

A removed resource whose destroy is blocked — protected by
`planton.dev/prevent-destroy`, its own deletion protection on, or still
read by anything else (another stack, a service, a resource deployed on
its own) or by a resource the templates declare — blocks only its
destroy: choosing Keep applies the plan around it. A Blocked
row on a resource the templates declare (another stack owns it, a secret
it reads is missing) stops the whole plan until it is resolved.

Beside the rows, a plan says:

- **Where the declaration came from**, above it: "Promoted from staging
  r6", "Roll back to r8", "Upgrade to app-infra v5", "Reconcile drift on
  orders-db and cache" (empty for an edit).
- **Where each value comes from**: set here, the chart's default, brought
  from another environment, or dropped (the template no longer declares
  it).
- **What Else This Affects**: everything that reads a resource the plan
  updates, replaces, destroys or may change — another stack, a service, a
  resource deployed on its own ("payments-infra in prod reads orders-db",
  "checkout-api, a service in prod, reads orders-db", "AWS Lambda 'notify'
  in prod reads orders-db"). Name each to the user by what it is. The
  readers themselves are not blocked, but a destroy of something one of
  them reads is: that row is Blocked, with Keep as the way through.
- **Its impact** (below): what the stack costs each month, the cloud
  permissions it adds, the security settings it changes, and the
  resources over the environment's deployment budget.
- **Template changes** and the **chart change** ("app-infra v4 → v5").
- **Warnings**: a template that names another environment literally, a
  direct edit in an environment that changes only by promotion. Nothing is
  blocked by a warning; relay it.

### The plan's impact

The platform works out what applying costs and what it opens up before
anything runs, from the same declaration the rows came from and the
catalog release's own cost, permission and control data. The CLI's
`plan`, `review` and `-o markdown` print it, `-o json` carries it as
`impact`, and the MCP plan and review tools return it as is. Read it
from there; never estimate it yourself.

- **Cost**: what the stack costs each month before and after, at the
  same prices, so the difference is the change: "$412 → $468 a month
  (+$56)", with each resource whose cost moves. The plan's cost sentence
  says what the figures are; relay it as it is. A figure that is only a
  range has no exact difference, and the plan prints none, so never
  subtract one. A resource the plan **keeps** leaves the stack's figures
  but keeps running and billing in the account, so it is listed apart
  ("legacy-queue keeps running outside it, $32 a month"): never present
  its departure as a saving.
- **Added permissions**: the cloud permissions the stack's resources need
  once the plan applies that they did not need before, by provider, each
  in the provider's own words (`rds:CreateDBInstance`,
  `container.clusters.create`) with the resources that need it. They are
  measured against what the stack needs today, never against what the
  runner's credential holds — the platform cannot read a cloud account's
  grants. Tell the user the runner's connection must allow them; a
  missing one fails that resource's job.
- **Security settings**: each setting the plan changes on a resource the
  stack already runs, named by its control ("Encryption at rest",
  "Network Exposure") with its value before and after. These are named,
  never graded: the catalog says which field sets a control, not which
  value is safer, so do not call a change safer or riskier. Deletion
  protection is the one the plan judges: "Turns Off Deletion Protection"
  is worth saying first.
- **Deployment budget**: the budget is per deployment, compared with each
  resource's own monthly figure. Each resource over it is listed ("2
  resources are over prod's $300 deployment budget"); each one's job
  pauses for an approver when it runs, so say so before the user agrees.

In the console the plan opens **View Architecture** (`F`): the stack's
graph as the run will walk it, each resource marked with what the plan
does to it, the unchanged ones dimmed and the destroyed or kept ones drawn
as ghosts. It is the run's own graph, so the picture is what applies. A
revision keeps its plan's impact and warnings, so an approver reads
exactly what the requester saw.

Exit codes: `0` planned; `1` the declaration does not build, a resource it
declares is blocked, or a run is in flight; `2` the plan could not be
made. With `--detailed-exitcode`, `4` means applying would change
something. `-o markdown` prints the plan for a pull request or an answer;
`-o json` for a program.

**2. Show.** Put the plan in front of the user in their words: what is
created, updated, replaced and destroyed — **every Destroy and Replace row
by name**, and that what a destroyed or replaced resource holds does not
come back — what the stack's monthly cost becomes, and any permissions it
adds, security settings it changes or resources over budget. A plan of the
declaration the stack already holds also has a page in the console (the
plan prints its link); a plan of an edited folder lives only in your
output.

**3. Consent.** Ask, naming each destroy and replace. The user's yes
covers every Destroy and Replace row the plan shows, so say each one
before they answer; nobody types or names a resource to consent to it.
If the user has told you to carry on without asking (for this task), that
is their agreement: apply, and afterwards tell them exactly what was
destroyed or replaced. Otherwise ask first.
Anything the user wants kept goes `--keep Kind/name`; an existing
resource the stack should take over goes `--adopt Kind/name` — a stack
never takes over a resource nobody chose to hand it. A kept resource keeps
what it depends on: keep a deployment and its namespace must be kept too,
because destroying the namespace would take the deployment along. Each
removed row lists the removed resources that depend on it
(`removed_dependents`); the apply refuses a keep without what it depends on,
naming every `--keep` it needs in one sentence. A Blocked removal applies
only with `--keep`. `--destroy Kind/name` does one thing: it turns a Keep
carried from an earlier revision back into a Destroy; on any other row it
is refused.

Two annotations on a resource's `metadata` settle the question before any
plan asks it; adding or removing either plans Update Record and runs no job:

```yaml
metadata:
  annotations:
    planton.dev/keep-on-provider: "true"   # never destroyed: whenever the stack lets go, it is kept
    planton.dev/prevent-destroy: "true"    # never torn down: its destroy plans Blocked, Keep applies around it
```

**4. Apply** exactly what was shown:

```
planton infra stack apply <stackName> . --env <env> --org <org> \
  -m "one line saying WHY, like a commit message" \
  --plan-digest <digest from the plan you showed> \
  --keep AwsRoute53Zone/legacy --plain
```

The digest is on the plan itself: its "Apply it:" line already carries
`--plan-digest <d>`, `-o markdown` prints it as "Plan digest:", and
`-o json` as `plan_digest`. In CI, post the `-o markdown` plan to the pull
request and apply with the digest of the plan a person approved there.
The apply plans again and goes ahead only while the fresh plan is the one
named by `--plan-digest` — otherwise it refuses, saying the plan is out of
date; plan again and show the new plan. `--plan-digest` and `--yes` are
both consent given in advance to every Destroy and Replace row at its
default; the digest also pins the exact plan the user read, so pass it.
At a terminal with neither, the apply shows the plan and asks once.
Without a terminal (your session, CI) and with neither, it refuses — "Add
--yes to Proceed Without Asking" — saying what the plan destroys and
replaces, with the command to run. Every consented apply prints what it
destroys and replaces before its run starts; relay it. The `-m` message is
mandatory practice: it names the revision and its run ("r13 · Raise
replicas"), where the whole team reads why it happened.

What one apply does: records the declaration and a numbered revision, then
starts ONE run that applies what the stack declares and — only when all
of it applied — removes what the plan removed (destroyed, or kept running
on its provider), prints the run's id, and **follows it to completion**; the exit
code is non-zero when the followed run fails. A plan that changes only the
declaration (a comment in a template, a value set to its default) records
the revision and starts no run. A plan that changes nothing applies
nothing: "No changes".

`planton chart install <stackName> . --env <env>` takes the same path (plan,
show, ask, apply), and creates the stack when the name is new; `--dry-run`
stops at the plan. Installed from a catalog chart (`planton chart install
<stackName> app-infra --env <env>`, or `org/<slug>`), the platform copies
the chart's files itself and records the chart and its version; onto a
stack already following that chart it is an upgrade
(`infra.stack-promotion.md`). A local chart folder records no chart. `planton apply -f <stack manifest>` applies a plan only
when nothing in it is destroyed, adopted or blocked; otherwise it refuses
and names the `planton infra stack apply … --yes` command to review it with.

A stack has one run at a time. A plan of a stack whose run is in flight
answers with that run instead of rows, and an apply is refused naming it
("… already has a run in progress (<id>) …", or "… has a request waiting
for approval: r14 · Raise replicas. Review it (planton infra stack review
…), or withdraw it (planton infra stack withdraw …) before asking for
another change."). Never retry blindly: tell the user, and let them decide
the waiting request, withdraw it, or cancel the run (`planton infra
pipeline cancel <id>`, a platform mutation — confirm first), then plan
again.

## In a protected environment: a request, not a write

In a protected environment the plan says **This Plan Needs an Approver**
(`review_required` in JSON), and its last line reads "Request it:" instead
of "Apply it:". Tell the user before they agree: applying files a REQUEST.
The engines preview what it changes, an approver reads that exact plan and
approves or rejects it, and only an approval applies — exactly the plan
read. Until then nothing applies and the stack's declaration does not move.

```
planton infra stack apply <stackName> . --env prod -m "<the user's words>" --plan-digest <d>
```

- **`-m` is required**, and it is the user's description: approvers read it
  first. Ask for it; never invent it.
- The command follows while the request previews, then **ends** — a
  terminal never waits on a person — with exit 0 and these lines. Relay
  them as they are; the review link is the user's next step:

  ```
  Waiting for approval: 1 to Replace, 3 to Update. Lapses at Oct 21, 14:30 IST, in 6 days.
  Review it: https://<console>/<org>/infra-stack/prod/<stackName>/plan?run=<infpipe_id>
  An approver runs: planton infra stack approve <stackName> --env prod
  Withdraw it: planton infra stack withdraw <stackName> --env prod
  ```

  `-o json` prints `{run_id, review_url, state, lapses_at, revision, source,
  from_env}` once the review opens. A request whose preview fails exits 1 with the engine's
  sentence. Undeploy, purge, `chart install|uninstall` and `infra pipeline
  rerun` into a protected environment are requests too, and end the same way.
  The one-call `planton apply -f` never files a request: it refuses and
  names `planton infra stack apply … --yes`.
- **Reading a waiting request** is free: `planton infra stack review
  <stackName> --env prod` (MCP: `get_infra_stack_review`). It lists what is
  destroyed and replaced first, with what each resource's preview found,
  the plan's impact and warnings exactly as the requester saw them, what
  else reads what it changes, who can approve and when it lapses, then
  the commands, and the review's digest last. Explain it in its rows' words: every Destroy and Replace by
  name, and that a **Replace** destroys and recreates the resource, so what
  it holds does not carry over.
- **Deciding is the user's act, never yours.** Never approve because a
  request is waiting, because the user wrote it, or on your own judgment.
  Carry out a decision when the user, having read the review, tells you to
  approve or reject — or approve when an approver told you to carry on
  without asking (for this task): that is their consent given in advance,
  and afterwards you tell them exactly what the approval destroyed or
  replaced. **Never approve a run your own person started**, even when
  they are its only approver: the platform refuses it through the agent
  stamp every call carries, and the refusal says the person can approve it
  themselves. Relay that.

  ```
  planton infra stack approve <stackName> --env prod --review-digest <d>
  planton infra stack reject <stackName> --env prod --review-digest <d> --reason "<the user's words>"
  ```

  (MCP: `decide_infra_stack_review` with `review_digest`, `confirm`, and a
  rejection's `reason`.) Approving approves exactly the review read, with
  every Destroy and Replace in it: nobody types or names a resource. At a
  terminal `approve` shows the review and asks; `--yes` alone prints the
  current review, then approves its digest; with neither and no terminal it
  refuses with the command carrying the current digest. A rejection always
  carries the user's reason; the requester reads it. The platform refuses
  a review that moved since it was read
  ("This review changed since you read it. Read it again …"), a blocked
  row, and the person who asked for the change while another approver
  exists — relay its sentence as it is.
- **After an approval**, `approve` follows the run: exit 0 when it
  succeeds, 1 when it fails, and **3** when a resource stopped before
  changing anything because its plan moved since the approval — the line
  above names `planton infra job resume <id>`, which is again the user's
  decision on the new plan.
- **Withdrawing** (`planton infra stack withdraw`, MCP
  `withdraw_infra_stack_request`) takes the request back; its requester may
  do it without deploy rights. Confirm first.
- **A request nobody decides within 7 days lapses.** It closes **Lapsed**
  ("Lapsed: no one decided within 7 days, so nothing was applied"), and the
  stack is free again. A request that is rejected, withdrawn or lapsed,
  fails its preview or goes out of date applies nothing. Plan it again as it
  was — its choices carried, each still the user's to change:

  ```
  planton infra stack plan <stackName> --env prod --revision 14
  ```

  Planning again keeps where the revision came from: a rejected promotion
  planned again is still a promotion ("Promoted from staging r6"), and a
  roll back or an upgrade likewise; a carried Keep keeps its origin ("kept
  in staging r6"). The new revision reads "Planned again from r14".

When holding one long-running command open is unsuitable, split it:

```
planton infra stack apply <stackName> . --env <env> -m "…" --plan-digest <d> --no-follow
planton infra pipeline status <infpipe_id>     # poll between other work
planton infra pipeline logs <infpipe_id>       # when a node fails
```

`planton infra stack history <stackName> --env <env>` lists the revisions,
newest first: number, message, source (an edit, a re-apply, "Promoted from
staging r6", "Rolled back to r8", "Upgraded to app-infra v5", "Reconciled
drift on orders-db and cache"), outcome (Applying, Applied, Applied With
Removal Failures, Failed, Cancelled; for a request Awaiting Approval, then
Rejected, Withdrawn, Lapsed, Preview Failed or Out of Date when it closed
without applying), who requested it and who approved it, and its run; the
live revision — the newest that applied — is marked.

## What a run applies, and what it leaves alone

A stack run applies only the resources whose inputs moved since their last
successful deploy — the same comparison the plan shows as Unchanged. An
exact match runs no job: the node reads **Unchanged since r11**, naming
the revision whose run last applied it (the CLI adds that job's id:
"Unchanged since r11 (ij_…)"), and hands that job's outputs to its
dependents. So a change to one value runs that resource, and a dependent
only when a value it reads changed; a re-apply with nothing changed runs
nothing. Say so when you narrate: unchanged nodes are successes, not
skipped work. `--reapply Kind/name` (or `--reapply-all`) runs a job anyway.

More node outcomes to read correctly:

- **Unmanaged** — a resource whose lifecycle events are disabled
  (`planton.dev/infra-job.disable-on-lifecycle-events: "true"`). The stack
  records it and never applies or destroys it: removing it from the
  templates plans Keep, and an undeploy releases it.
- **Kept** — a removed resource the user chose to keep, or one
  marked `planton.dev/keep-on-provider: "true"`: the stack let go of it, no job
  ran, and it keeps running; another stack or a service can adopt it.
- **Owned Elsewhere** — a component another live stack or service owns.
  A teardown leaves it in place for its owner; a plan that would change it
  is Blocked, naming the owner. Change it through that owner.
- **A resource changed outside Planton** (edited in the cloud console,
  deleted by hand) is invisible to that comparison, because nothing
  Planton deploys moved — a plain re-plan reads it Unchanged until a drift
  check finds it. Check for drift, then Reconcile (next section).

A run whose declared resources applied but a removal failed ends
**"Applied; 1 removal failed"**: what the stack declares is live, and the
revision reads Applied With Removal Failures. Diagnose the failed destroy
like any failed node.

## Changed outside Planton: check for drift, then Reconcile

Someone edited a database in the cloud console during an incident, or
deleted a bucket by hand. A **drift check** finds it: for each resource,
it compares what Planton recorded when it last applied it with what is
really in the cloud, setting by setting, and writes nothing — not the
state, not the resource. Checking is free to do; it changes nothing and
never waits for an approver, not even in a protected environment.

```
planton infra stack drift check <stackName> --env <env> --org <org>   # every resource the stack holds
planton infra stack drift show <stackName> --env <env> --org <org>    # where each resource stands
planton infra component drift check <ic_id>                           # one resource (or: <Kind> <name>)
planton infra component drift show <ic_id>
```

`check` follows until every resource has its verdict; `show` reads the
latest verdicts without checking again. Both take `-o table|json|markdown`.
Exit codes: `0` checked, `1` a resource could not be checked, `2` the
check could not run; with `--detailed-exitcode`, `4` means drift was
found. Agents: `check_infra_stack_drift` / `get_infra_stack_drift`, and
`check_infra_component_drift` / `get_infra_component_drift` for one
resource.

Each resource reads one verdict:

- **In Sync** — the cloud matches what Planton applied.
- **Drifted · 3 changes** — something was changed outside Planton. The
  change list names each setting as **Applied by Planton** and **In the
  Cloud Now** ("instance_class · db.t3.medium → db.t3.large"). A secret
  reads "Hidden value changed", never its value; a resource removed by
  hand reads "Deleted Outside Planton".
- **Couldn't Check** — with the reason ("The drift check failed: …", or
  "The runner did not run the drift check; it may need updating."). It
  never erases an earlier Drifted verdict: a check that could not look
  proves nothing.

A resource Planton never deployed has nothing to compare with and is
refused ("orders-db was never deployed by Planton, so there is nothing to
compare the cloud with. Deploy it first."); a stack check skips it.

**What drift does next.** Once a check finds drift, the next plan of the
stack — any plan, even an unrelated edit — includes that resource as
Update, reason "Drifted: 3 changes made outside Planton", and applying
puts the declaration back. Say so when you show any plan with such a row:
the user may have made that cloud change on purpose.

**Ask which they want** — this is the user's decision, never yours:

- **Put it back: Reconcile.** It plans the declaration the stack already
  holds and names what it reverts ("Reconcile drift on orders-db and
  cache"). Then show, ask and apply exactly as for any plan above — in a
  protected environment it files a request:

  ```
  planton infra stack reconcile <stackName> --env <env> --org <org>
  planton infra stack reconcile <stackName> --env <env> --org <org> -m "<why>" --plan-digest <d>
  ```

  The first prints the plan and, at a terminal, asks once; the second
  applies exactly the plan the user read. Reconcile takes everything
  `apply` takes (`--plan-digest`, `--yes`, `--destroy`, `--keep`,
  `--adopt`, `--no-follow`) plus `--dry-run`, which stops at the plan.

  A stack is never applied in part, so a resource that also changes for
  another reason (a newer module, a new secret version) stays in the plan,
  with a warning naming it — relay it. With nothing drifted it is refused
  ("Nothing in payments-infra in prod has drifted since its last check.
  Check for Drift to look again."). Agents plan it with `plan_infra_stack`
  and its `reconcile` source. A resource deployed on its own is put back by
  deploying it again; a service's resource by redeploying the service.
- **Keep the cloud's change: edit the template.** Carry each changed value
  into the template file that declares the resource (the change list names
  the setting and its value in the cloud), then plan and apply as usual;
  the next plan has nothing left to put back. In the console, **Edit in
  Template** opens that file.

A successful apply clears the resource's Drifted verdict. A plan alone —
`planton tofu plan`, `planton pulumi preview` — is not a drift check: it
answers "what would a deploy change?", not "does the cloud still match?".

## Narrate every transition — the user is watching the app

The app watches the stack's runs: the moment your apply starts one, the
screen announces it and navigates to the run. Say what you did the moment
you did it, so the voice and the screen agree:

- After the apply: say which revision was recorded and that its run is
  running — and that they are being taken to it.
- On success: say it plainly, summarize what changed.
- On failure: diagnose again (same four steps), explain, recommend, ask.

## When the fix does not take

A correct fix can still fail with the SAME error as before. Verified failure
mode: an Infra Component whose deploy failed can be left in a failed state
that keeps feeding the old configuration to new runs — the stack's record
and even the component's record show your corrected values, yet the run
fails with the stale ones. Recognize it by the signature (identical error
after an apply that provably changed the relevant field), then clear it:

```
planton search infra-components --org <org> -e <env>      # find the component id
planton get infra-component <ic_id> -o yaml               # confirm the record already carries your fix
planton infra component cleanup <ic_id>                   # clear the failed state (confirm with the user first)
```

Then rerun the run (`planton infra pipeline rerun <infpipe_id>`, or Rerun
on its page in the console): a rerun repeats the same revision, so what
applied stays Unchanged and what failed runs again. It asks first only when
the rerun still destroys or replaces something (`--yes` answers, once the
user agreed); agents use `rerun_infra_pipeline`, with `confirm` on the same
condition. A
rerun is refused once a newer revision exists or the stack was undeployed
since — plan again and apply instead; the refusal names the plan command.
A followed run that fails ends by naming the way out that fits it. Cleanup only clears the failed
deployment state — it does not touch cloud infrastructure — but it is
still a platform mutation: say what it does and get a yes.

A different non-taking failure has its own playbook: a rerun that fails
because a provider resource **already exists** (an earlier run created it
on its provider, then died before the state recorded it). Do not
destroy-and-retry — import it into state with the platform's import commands. Read
`infra.state-import.md`.

## Iterating until green

Failures can be layered. After the first fix attempt, ask ONE question that
sets the collaboration mode for the rest of the session:

> "Want me to keep iterating until this deploys green — showing you each
> plan before I apply it — or check in with you after every diagnosis?"

Honor the answer. Even in keep-iterating mode, every apply still shows its
plan (what will change and why, one line each, every destroy and replace by
name) and gets its own yes, unless the user told you to carry on without
asking: that is their agreement, so apply, and afterwards tell them exactly
what was destroyed or replaced. What the mode changes is whether you pause
after each diagnosis or proceed to the proposed fix's plan directly.

## Boundaries

- Never apply without the user's go-ahead for that specific plan, or their
  word to carry on without asking (for this task); after a carried-on apply,
  say exactly what was destroyed or replaced.
- Never change the target env/org to "make it work" — a failure that smells
  like wrong-environment is a finding to report, not to route around.
- Undeploying, purging, or deleting the stack is the user's to drive in the
  product; surface the need. What to tell them when they ask: undeploy
  destroys everything the stack owns, including components its templates
  no longer declare, dependents first, and keeps the stack and its
  revisions; unmanaged resources and ones marked `planton.dev/keep-on-provider`
  are released and stay on their provider. It is refused while any owned
  resource carries `planton.dev/prevent-destroy: "true"` (lift it by
  removing the annotation and applying) or while anything else still reads
  one of its resources — another stack, a service, a resource deployed on
  its own (the refusal names the reader). Applying the stack again
  later plans each resource anew — any the undeploy released plans Adopt.
  Purge is undeploy, then delete, only if the undeploy succeeded. Delete
  alone is refused while the stack still owns components or has a run in
  flight. From a session, each of these runs only with `--yes`, and only
  after the user said yes or told you to carry on without asking; then say
  exactly what was destroyed.
