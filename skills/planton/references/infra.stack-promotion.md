# Promote, Roll Back, Upgrade, Compare — A Stack Across Its Environments

A stack lives on its environment's Promotion Path, and its identity is its
name on that path: **the same name on the same path is the same stack.**
`app-infra` in qa, staging and prod on the homes path is three separate
records — each with its own values, revisions, runs and access — and one
stack at three stages. `app-infra` in an environment on another path, or on
no path, is another stack. A template moves between stages of its path by
**Promote**; each environment keeps the values it set. Every apply is a
revision you can return to by **Roll Back**. A stack installed from a
catalog chart follows that chart, and **Upgrade** moves it to the chart's
newest version while keeping its values. **Compare** puts two environments,
or two revisions, side by side.

Promote, Roll Back and Upgrade are not new kinds of run. Each is a **plan
source**: it builds the declaration to plan, then ends in the same plan,
show, consent and apply as every other change (`infra.deployed-stacks.md`).
Everything that reference says about plans, rows, digests, destroys, Keep,
Adopt, requests in a protected environment and one run at a time holds here
unchanged. Read it first.

## Where a stack stands: the Delivery Path

```
planton infra stack status <slug>              # writes nothing
planton infra stack status <slug> -o json      # the path itself (or -o yaml)
```

One section per Promotion Path the name lives on, headed by the path's name
and where Promotion Only starts; one record per environment with its stage
number (siblings share it): what runs there on the first line, its chart
version and how far it is from its comparison source on the second, then
each run moving in. The environments on no path that hold the name follow,
then the commands that move it on:

```
Status · app-infra

homes · Promotion Only from stage 3
1  qa                   r9 · Raise replicas
                        app-infra v4
2  staging              r6 · Raise replicas
                        app-infra v4 · Same template as qa
3  prod-eu (protected)  r14 · Pin version · Requested r15
                        app-infra v3 · 2 revisions behind staging
                        Promoting from staging · Awaiting Your Approval · r15
   prod-us (protected)  Not Applied · Promote from staging would create it

On No Promotion Path (no promotion leads here)
   sandbox              r2 · Try it
                        app-infra v4

r15 waits for your approval in prod-eu: planton infra stack review app-infra --env prod-eu
Promote staging to prod-us: planton infra stack promote app-infra --from staging --to prod-us
```

- **What runs**: the live revision ("r9 · Raise replicas") and the chart
  version it runs, then a newer revision that failed ("r10 Failed") or a
  request waiting ("Requested r15"). **Not Applied** when the environment
  has nothing applied (never applied, or undeployed since); a first request
  that closed without applying reads "Not Applied · Rejected" (the stack
  page: "Not Applied · Request Rejected"). Where Promote would create the
  stack, it says so ("Promote from staging would create it").
- **The distance** from its comparison source: the earlier-stage environment
  its live template was promoted from, else the nearest earlier stage's
  newest — never a sibling, never another path. The first stage holding the
  stack compares with nothing. It is read from the revisions' template
  digests, never from timestamps — so a roll back is never mistaken for
  being behind:
  - **Same template as qa** — byte-equal templates and `values.yaml`.
  - **2 revisions behind staging** — staging moved past this template;
    Promote brings it here.
  - **Ahead of qa** — this environment runs a template qa moved away from,
    or never had yet. Promoting from qa would take it back.
  - **Differs from staging** — neither history holds the other's template:
    this environment was changed on its own.
- **Update Available** — the chart has a newer version than the
  environment runs ("app-infra v4 · v5 available"). **customised**
  ("app-infra v4 · customised") — the stack's files differ from that chart
  version's.
- **What is moving in** — a run applying, a request previewing or waiting
  for approval ("Promoting from staging · Applying · r16").
- An environment whose stack you cannot read says only that ("You can't
  read app-infra in prod"). Environments on no path follow under **On No
  Promotion Path**, each a stack of its own, never measured against another.

The path is the answer to "is prod behind?", "did staging get the latest
chart version?" and "what's waiting?". Agents read the same answer with
`get_infra_stack_delivery_path` and say it in the same words.

## Promote

```
planton infra stack promote <slug> --from <env> [--to <env>] \
  [--bring <name>]... [--set name=value]... [-m "<message>"]
```

`--to` may be left out: it is then the next stage after `--from` on the
stack's Delivery Path — its one environment, or at a terminal the sibling
the person chooses; without a terminal, a stage of several is refused with
the command to rerun with `--to`. Each sibling is promoted on its own.

What travels and what stays:

- **Travels**: the source's live template, its `values.yaml` and its chart
  version ("Template from staging r6 · app-infra v4"). Templates only ever
  come from the source's LIVE revision.
- **Stays**: every value the target environment set. A value travels only
  when named with `--bring <name>` (the console's **Bring This Value**).
- **Needs a value**: a param the travelling `values.yaml` newly requires
  with no default. The plan names it ("Set a value for `alert_email`");
  pass `--set alert_email=…`.
- **Dropped**: a value the target set for a param the travelling
  `values.yaml` no longer declares. The plan shows it as dropped.
- **Kept on the provider**: every resource the source keeps running after
  removing it from its templates is offered as Keep in the target too, and
  reads "kept in staging r6" — a promotion never silently destroys what the
  source chose to keep. A source Keep with no match in the target is a plan
  warning ("legacy-queue was kept in staging; prod has none"). A person's
  Destroy, Adopt or Re-apply in one environment never carries to another.

More of what to know:

- **A first promotion** into an environment with no stack of that name
  creates it there, with every value from the source.
- **Either direction works** along the path: "make qa like prod" is
  `--from prod --to qa`. The path draws steps forward only.
- **A promotion stays on its Promotion Path.** Across two paths, or into or
  out of an environment on no path, it is refused, naming both places and
  the apply that reaches any environment directly ("A promotion stays on its
  Promotion Path: 'prod-eu' is on the homes path and 'zm-staging' is on the
  mortgage path. ...").
- **The message** defaults to the source revision's message ("Raise
  replicas"); the revision's source line reads **"Promoted from staging
  r6"**, and so do its run's page, the activity feed and notifications.
- **A template that names its source environment literally** is a plan
  warning, never a block: "templates/db.yaml names qa (line 12). It reads
  the same in prod. Use {{ values.env }}." Tell the user; fix the chart if
  they agree.
- **Into a protected environment** the promotion is a request: the approver
  reads "Promoted from staging r6" above the exact plan, the losses first,
  its cost and added permissions, and what else it affects. `-m` is
  required, as for any request.
- **In a Promotion Only stage**, a template edit that did not come by
  promotion is cautioned on its plan when the stack also lives in an earlier
  stage of that path ("prod changes by promotion from staging. This edit
  makes prod differ from staging."), and the path reads "Differs from
  staging". It is never refused; protection decides who may change the
  environment.
- Refused, with the sentence to relay: promoting into the same environment,
  a source environment with no stack of that name, a source with nothing
  live, and an environment being deleted. Promoting needs read access on
  the source stack and deploy access on the target.

## Roll Back

```
planton infra stack rollback <slug> --env <env> [--to-revision N] [--adopt Kind/name]...
```

Roll Back plans an earlier revision's declaration — its templates,
`values.yaml`, values and chart version — fresh against today's cloud, as
a new revision ("Roll back to r8"; its source line reads "Rolled back to
r8"). `--env` is always explicit.

- **The default target** is the live revision when the latest one failed,
  otherwise the newest applied revision before the live one. Pass
  `--to-revision N` when the user named one; the target must have applied
  (Applied, or Applied With Removal Failures).
- **Nothing its author chose then is carried**: those choices were made
  against a different stack. The plan decides each resource again.
- **Say these before the user agrees** — they are the dialog's own
  caution:
  - **Module versions stay current.** A roll back restores the
    declaration, not the modules that applied it.
  - **Data destroyed since comes back empty.** A database destroyed after
    that revision is recreated, not restored.
  - A resource kept on its provider since then plans **Adopt**; name it
    with `--adopt Kind/name` to take it back.
  - A resource another stack now owns plans **Blocked**, naming the owner.
- A roll back to a revision with no chart record clears the stack's chart
  record, and the plan says so; one to another chart version brings Update
  Available back.
- Refused while a request waits or a run is in flight on the stack — decide
  or withdraw the request, or let the run end, then plan again.

## Upgrade

```
planton infra stack upgrade <slug> --env <env> \
  [--keep-mine <file>]... [--take-chart <file>]...
```

Upgrade moves a stack to its catalog chart's newest version ("Upgrade to
app-infra v5") and keeps its values. Only a stack installed from a catalog
chart has one; a stack planned from a local chart folder follows no chart.

Each file is compared three ways — the version the stack is on (kept by its
own revisions), the stack's own file, and the new version's:

| The file | What the plan does |
|----------|--------------------|
| Unchanged by you | Takes the new version's file. |
| Changed only by you | Keeps yours. |
| Changed by both | A conflict: the plan cannot apply until you choose **Keep Mine** or **Take v5** (`--keep-mine <file>` / `--take-chart <file>`). |
| Deleted by the new version | A change: Take v5 removes it, and its resources plan Destroy. |

- `values.yaml` is a file like any template (the key `values.yaml`).
  Template keys are the stack's bare file names; `--keep-mine` accepts them
  with or without `templates/`.
- A value set for a param the new version no longer declares reads
  "Dropped: v5 no longer declares it". A newly required param asks for a
  value (`--set`).
- An upgrade whose files already match the new version records a revision
  and starts no run.
- On a later stage whose earlier stage already runs the new version, the
  path reads "qa runs v5 · Promote from qa": promote the tested template
  instead of upgrading each environment on its own.
- `planton chart install <name> <catalog-chart> --env <env>` onto a stack
  that already follows that chart is an upgrade; a conflict there names
  `planton infra stack upgrade` with its flags.

## Compare

```
planton infra stack diff <slug> --env <env> --against <env>
planton infra stack diff <slug> --env <env> --revision N --against-revision M
```

Two environments (each side's live revision), or two revisions of one
environment. Values first, each side with where it comes from (set here,
chart default); then how far apart the templates are; then the resources
only one side declares; then the files that differ. Hidden values are
redacted. Exit `0` when the two are the same, `1` when they differ, like
`planton diff`. Agents use `diff_infra_stack`. In the console, **Make prod
Like qa** opens Promote.

## What Promote, Roll Back, Upgrade and Reconcile all take

- The plan, show, consent and apply steps of `planton infra stack apply`:
  at a terminal it shows the plan, says what it destroys and replaces, and
  asks once; `--plan-digest <digest>` applies only the plan a person
  reviewed, and `--yes` applies the fresh plan; either consents to every
  Destroy and Replace it shows, and the apply prints what it destroyed and
  replaced. `--keep` and `--adopt` name choices; `--destroy` only turns a
  carried Keep back into a Destroy. Without a terminal and with neither,
  it refuses and prints the command with `--yes` added.
- `--dry-run` stops at the plan. `--detailed-exitcode` adds exit `4` when
  applying would change something. `-o json|markdown|yaml` prints the plan
  for a program, a pull request or a record.
- Exit codes as for `plan`: `0` planned (or applied), `1` it does not
  build, a declared resource is blocked or a run is in flight, `2` the
  plan could not be made.

**CI flow** — the plan goes to the pull request, and only the reviewed plan
applies:

```
planton infra stack promote app-infra --from staging --to prod --dry-run -o markdown   # post to the PR
planton infra stack promote app-infra --from staging --to prod -m "Raise replicas" \
  --plan-digest <digest from that reviewed plan>
```

If anything moved since the plan was posted, the apply refuses ("This plan
is out of date…"); post the new plan.

## The plan's parts these sources add

- **The source line** above the plan: "Promoted from staging r6", "Roll
  back to r8", "Upgrade to app-infra v5".
- **Where each value comes from**: set here, the chart's default, brought
  from staging, or dropped.
- **What Else This Affects**: everything that reads a resource the plan
  updates, replaces, destroys or may change — another stack, a service, a
  resource deployed on its own ("payments-infra in prod reads orders-db",
  "checkout-api in prod reads orders-db"). Name each to the user. A plan
  that would destroy something one of them reads is Blocked on that row,
  with Keep as the way through.
- **The plan's impact** (cost, added permissions, security settings, the
  deployment budget), as for any plan (`infra.deployed-stacks.md`). A
  promotion that moves a chart version often moves the cost too: say it.
- **Template changes**: each file added, removed or changed, and on an
  upgrade how it was decided.
- **The chart change**: "app-infra v4 → v5", or a different chart, with a
  caution.
- **Warnings**: an environment named in a template, a source Keep with no
  match, a direct edit in a promotion-only environment.

## Agents

`plan_infra_stack` and `apply_infra_stack_plan` take the same sources: a
promotion (from an environment, values to bring, values to set), a roll
back (a revision), or a chart (its file choices). `get_infra_stack_delivery_path`
answers where the stack stands; `diff_infra_stack` compares.

**Consent is the same as any apply.** Never promote, roll back or upgrade
because a stage is behind, because Update Available shows, or because a
run failed. Show the plan in the user's words — what travels, what stays,
every Destroy and Replace by name, its cost, what else it affects, the
roll back's caveats — and apply only on their yes for that plan, or under
their word to carry on without asking (then say exactly what was destroyed
or replaced). Into a protected environment the apply files a request;
deciding it is a person's act, carried out only on their word and never on
a run your own person started (`infra.deployed-stacks.md`).
