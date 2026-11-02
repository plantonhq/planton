---
title: "Infra Stacks"
sidebar_title: "Infra Stacks"
description: "An Infra Chart declared into one environment: every change is a plan you read first, and every applied plan is a numbered revision."
icon: infrastructure
order: 40
tags:
  - Infrastructure
  - Infra Stacks
  - Deployment
---

# Infra Stacks

An Infra Chart defines what you *can* deploy — a reusable template with parameterized placeholders. But a template alone does not answer the questions that matter for a specific deployment: which region, which domain name, how many availability zones, which environment. An Infra Stack captures those answers. It is a template declared into one environment with the values set there, applied as real infrastructure -- and every change to it is a plan you read before anything runs.

The relationship mirrors how Helm works: an Infra Chart is to an Infra Stack as a Helm chart is to a Helm release. The chart is reusable; the stack is the deployed instance with specific values. This separation enables teams to capture infrastructure patterns once and instantiate them many times — across different environments, regions, or customers — each instance tracked as a distinct Infra Stack.

## Why Infra Stacks Exist

An Infra Chart gives reusability, but a chart alone keeps no record of each place it was deployed. Consider an organization that deploys five ECS environments from the same chart — development, staging, production-us, production-eu, and production-asia. Each uses different parameter values. Without a stack record, fundamental questions have no answers:

- Which parameters deployed production-us last month?
- Did staging update to the latest chart version?
- Who changed the load balancer domain in production-eu?

Infra Stacks answer them with a persistent record of every template instantiation. Each stack holds its declaration — the chart's templates, the chart's `values.yaml` (every value declared, with its default) and the values set for its environment — with its render, its dependency graph, and a numbered **revision** for every plan applied to it: who applied it, why, and how its run ended. Runs are transient; the stack and its revisions are the source of truth for what was declared and when.

## One Chart, One Environment

You select an Infra Chart from the catalog, pick the environment it deploys into, set the values that environment needs, and review the plan of what installing creates.

The stack carries its own copy of the chart's templates and `values.yaml`. Chart updates never change an existing stack silently — each stack is self-contained, and it can be checked out to disk and edited with no chart in reach. A stack installed from a catalog chart also records which chart and which version it came from ("app-infra v3"). That record is the platform's, kept beside the declaration rather than inside it, so no file can claim a chart it did not come from, and every later edit keeps it. When the chart publishes a newer version, the stack reads **Update Available**, and [Upgrade](#upgrade) brings it in. Only the values someone set are stored as the environment's values; every other value takes the default from the stack's `values.yaml`, so a default the stack's `values.yaml` changes reaches every environment that never chose otherwise — and the plan shows it as a value change first.

The web console provides a creation flow: browse the chart catalog, pick the environment, fill in the values form (generated from the chart's declarations, each default shown as a placeholder), then **Review Plan**.

<!-- SCREENSHOT: Infra Stack creation from chart
  Page: /[org]/infra-stacks/create/[infraChartId]
  Action: Show the creation wizard with chart parameters form
  Focus: The parameter form and environment selector
  Alt: Infra Stack creation wizard showing chart parameters and environment selection for an AWS ECS environment chart
-->

## Every Change Is a Plan

Every change to an Infra Stack — an edit in the console, `planton infra stack apply`, `planton chart install`, `planton apply -f`, an agent's tool — is first a **plan**: what applying the declaration would do to each resource, before anything runs.

| Row | What applying does |
|-----|--------------------|
| **Blocked** | Nothing, until it is resolved; the row says why and the way out. |
| **Destroy** | The resource is no longer in the templates: it is destroyed on its provider once everything the plan applies has applied. |
| **Replace** | The resource stays in the templates, but its provider can't change it in place: it is destroyed and created again, and what it holds does not carry over. A request's preview finds it; it needs the requester's destroy rights, or the row reads Blocked. |
| **Update** | Something that decides its deployment moved: its settings, a secret's version, a value it reads, its module. |
| **Update Record** | Only Planton's record changes; nothing on the provider (for example, an annotation added). |
| **Create** | The resource is new. |
| **Adopt** | The resource already exists and no stack manages it; the stack takes it over only when you choose Adopt, and the apply may change it to match the template. |
| **Keep** (named by its provider: Keep in AWS) | The resource is no longer in the templates and is kept: Planton stops managing it, and it keeps running where it is, and billing there. |
| **May Change** | It reads a value known only after an earlier resource applies. |
| **Always Runs** | Its job runs on every apply: whether it changes is known only when it runs. |
| **Unchanged** | Nothing it depends on moved since its last successful deploy. |
| **Unmanaged** | Planton records it and never applies it. |

Applying a plan runs exactly that plan, and only while nothing it was made from has moved: the apply plans again and is refused when the fresh plan differs ("This plan is out of date…") — someone applied in the meantime, or a resource changed underneath. Plan again and review what changed.

An apply records a numbered **revision** named by its message ("r13 · Raise replicas") and starts one [Infra Pipeline](/docs/infrastructure/infra-pipelines) run: it applies what the stack declares, each resource after everything it consumes, up to four at once; then — only when all of it applied — it removes what the plan removed, dependents first: destroyed, or let go of and kept on its provider. A failed apply removes nothing. A plan that changes only the declaration (a comment in a template, a value set to its default) records its revision and starts no run; a plan that changes nothing applies nothing.

### Choosing What Happens to Each Resource

- **Destroy or Keep.** A removed resource is destroyed unless you keep it (**Keep** in the console, `--keep Kind/name` on the CLI). Nothing is typed to destroy it: applying the plan you read is the consent. A removed resource whose destroy is blocked — `planton.dev/prevent-destroy`, its own deletion protection, or something that still reads it — blocks only its destroy: choose Keep and the plan applies.
- **Adopt.** A stack never takes over a resource it found; you choose Adopt for it, or remove it from the templates.
- **Re-apply.** Run an unchanged resource's job anyway, or every resource's (`--reapply Kind/name`, `--reapply-all`).

### What a Plan Costs and Opens Up

Before anything runs, a plan says what applying it costs and what it opens up. Planton works it out from the same declaration the rows came from and the pinned catalog release's own cost, permission and security data, so the console, the phone, `planton infra stack plan` and an agent all read the same figures:

- **Cost.** What the stack costs each month before and after, at the same prices, so the difference is the change: "$412 → $468 a month (+$56)", with each resource whose cost moves. A figure known only as a range shows no exact difference. A resource you keep leaves the stack's figures but keeps running and billing in your account, so the plan says so instead of counting it as a saving: "$412 → $380 a month for this stack; legacy-queue keeps running outside it, $32 a month".
- **Added Permissions.** The cloud permissions the stack's resources need once the plan applies that they did not need before, by provider and in the provider's own words (`rds:CreateDBInstance`, `container.clusters.create`), each with the resources that need it: "Adds 4 Permissions This Stack Didn't Need Before". They are measured against what the stack needs today, never against what your runner's credential holds, because Planton cannot read a cloud account's grants. The runner's connection must allow them.
- **Security Settings.** Each security or resilience setting the plan changes on a resource the stack already runs, named with its value before and after: "Changes 2 Security Settings: Network Exposure, Encryption at Rest". Settings are named, never graded: the catalog says which field sets a control, not which value is safer. Deletion protection is the exception, where the plan says the direction: "Turns Off Deletion Protection".
- **Deployment budget.** Where an environment sets a deployment budget, each resource whose own monthly cost is over it is named: "2 resources are over prod's $300 deployment budget; each pauses for an approver when it runs".

In the console the plan's summary shows the cost and two chips, **Added Permissions** and **Security Settings**, each opening the full list. `planton infra stack plan`, `review` and `-o markdown` print the same sections, so a pull request carries them too, and `-o json` carries them as `impact`.

**View Architecture** (`F`) draws the plan as the architecture it changes: the stack's graph exactly as the run will walk it, each resource marked with what the plan does to it, the unchanged ones dimmed and the destroyed or kept ones drawn as ghosts. It is the run's own graph, so the picture is what applies. Click a resource to go to its row.

`planton apply -f` applies a plan only when nothing in it is destroyed, adopted or blocked; otherwise it refuses and names the `planton infra stack apply … --yes` command to review the plan with. It never files a request: in a protected environment it refuses and names that command.

### In a Protected Environment, a Change Is a Request

In a protected environment applying a plan files a **request** instead of writing anything. The plan says **This Plan Needs an Approver**, and the request carries a message its approvers read first:

```bash
planton infra stack apply app-infra --env prod -m "Raise replicas"
```

The request's run previews what the change reaches — each preview shows its engine's plan and changes nothing — and an approver reads that exact plan, what is destroyed and replaced first, in the console, on the phone, or in a terminal. The revision keeps its plan's cost, permissions, security settings and warnings, so the approver reads exactly what the requester saw, along with everything that reads what the change touches:

```bash
planton infra stack review app-infra --env prod
planton infra stack approve app-infra --env prod
planton infra stack reject app-infra --env prod --reason "Not during the freeze"
planton infra stack withdraw app-infra --env prod   # the requester takes it back
```

- **Nothing applies until an approval**, and the stack's declaration does not move: a rejected request leaves only its revision in history.
- **The approver approves exactly the plan they read.** Each resource plans again and applies only when its plan matches its preview; one that moved since — a setting edited in the cloud, a secret's new version — stops before changing anything, shows its new plan and waits for a person. A resource that could only be previewed once an earlier change applied (May Change) applies under the approval unless its plan destroys or replaces something.
- **One approval decides the whole plan.** Approving applies every Destroy and Replace the review shows. At a terminal, `approve` shows the review and asks; without one, `--review-digest <digest>` approves exactly the review read, and `--yes` prints the current review, then approves it.
- **A second pair of eyes.** The person who asked cannot approve their own change while another approver exists. An agent acting for an approver may approve when the person told it to carry on without asking, and then says what the approval destroyed or replaced; it never approves a change its own person asked for, even when that person is the only approver.
- **A request nobody decides lapses.** While it waits it reads "Lapses at Oct 21, 14:30 IST, in 6 days"; after 7 days undecided it closes **Lapsed**: no one decided within 7 days, so nothing was applied, and the stack is free for the next change.
- **A terminal never waits on a person.** The apply follows while the request previews, then ends with where to review it and the commands that decide it; `-o json` prints `{run_id, review_url, state, lapses_at, revision, source, from_env}`.

A request that is rejected, withdrawn or lapsed, cannot be previewed or goes out of date applies nothing. Plan it again as it was, with the choices its author made: `planton infra stack plan app-infra --env prod --revision 14`. Planning again keeps where it came from: a rejected promotion planned again is still a promotion, and its new revision reads "Planned again from r14". Undeploy, purge and rerun into a protected environment are requests too.

### Revisions

Each applied plan is a revision: its number, message, author, where its declaration came from (an edit, a re-apply, "Promoted from staging r6", "Rolled back to r8", "Upgraded to app-infra v5", "Reconciled drift on orders-db and cache"), the chart version it took, the declaration it applied, the plan's rows as reviewed and chosen, and how its run ended — **Applying**, **Applied**, **Applied With Removal Failures** (what the stack declares applied; a removal failed), **Failed** or **Cancelled**. A request reads **Awaiting Approval** until someone decides it, and closes **Rejected**, **Withdrawn**, **Lapsed**, **Preview Failed** or **Out of Date** when it applies nothing. The stack's **live revision** is the newest that applied; an undeploy clears it. The stack page's **Revisions** tab and `planton infra stack history <name> --env <env>` list them, newest first, with each one's source and who requested and who approved it; from a revision you can [Roll Back](#roll-back) to it or [Compare](#compare) it with what is live.

A run's Rerun (`planton infra pipeline rerun <pipeline-id>` in a terminal) repeats its revision — the same reviewed plan — so what applied stays Unchanged and what failed runs again. It asks first when the rerun still destroys or replaces something; `--yes` answers ahead. A rerun is refused once a newer revision exists, or after an undeploy: plan again instead.

### One Run at a Time

A stack has one run at a time. While a run is queued, running or waiting for approval, a plan answers with that run instead of rows, and an apply, undeploy, purge or rerun is refused, naming the run — or the request waiting for approval, with the commands to review or withdraw it. Cancel a run with `planton infra pipeline cancel <pipeline-id>`, then try again. A run whose workflow is gone is ended by the platform within about ten minutes, so a stack is never left busy.

## Dependency Graph Visualization

A stack's page lists its resources first: each one's name, kind and the revision that last applied it ("orders-db · AWS RDS Instance · Applied in r9"), searchable and filterable by kind. **Remove from Stack** on a row opens the plan of the stack without it, where you choose Destroy or Keep; when the resource shares its template file with others, it opens that file in the stack's configuration instead, saying which document to remove.

Behind **List | Graph**, each Infra Stack maintains a dependency graph that visualizes the relationships between its Infra Components. The web console displays this as an interactive DAG where you can see:

- Resource nodes with their current deployment status (queued, running, succeeded, failed), and, for a resource the run left as it was with no job, its outcome: **Unchanged since r11**, **Unmanaged**, **Kept** (named by its provider: Kept in AWS), **Owned Elsewhere** or **Already Gone**
- Dependency edges showing which resources must be deployed before others
- Real-time updates as a pipeline progresses through the graph

This visualization makes it easy to understand the deployment topology and diagnose failures — if a resource fails, you can immediately see which downstream resources are blocked.

<!-- SCREENSHOT: Infra Stack DAG visualization
  Page: /[org]/infra-stack/[infraStackSlug]
  Action: Show the Infra Stack detail page with the DAG tab active
  Focus: The dependency graph visualization showing resources and their connections
  Alt: Infra Stack DAG visualization showing an AWS VPC connected to subnets, security groups, and an RDS instance
-->

## One Stack, Many Environments

An Infra Stack lives in exactly one environment, and its name is unique within that environment. A stack is its name on its [Promotion Path](/docs/ci-cd/deployment-environments#promotion-paths-the-order-a-push-walks): the same name in another environment of the same path is the same stack at another stage. `app-infra` in qa, staging and prod on one path is three records — each with its own values, its own history and its own runs, and changing one never touches another — and one stack moving through its stages. The same name on another path, or in an environment on no path, is another stack. A template moves between stages by [Promote](#promote), and each environment keeps the values it set. The Infra Stacks page shows one card per stack — one per name on each path, and one for each environment on no path that holds the name — with a line for each of its environments in the order a push walks them. Every command that names a stack by name also names its environment with `--env`; a stack's id identifies it on its own. A stack's environment is fixed when it is created.

The environment also decides who may change the stack. Stacks follow the same access rules as the Infra Components they deploy: a permission granted on an environment covers that environment's stacks, and a protected environment stops organization-wide deploy and destroy permissions at its boundary.

### Where Each Environment Stands

*Did staging update to the latest chart version?* The stack's page opens on its **Delivery Path**, headed by its Promotion Path's name: its environments stage by stage, siblings side by side, each saying what it runs and how far it is from the environment it is compared with — the one its template was promoted from when that sits in an earlier stage, otherwise the nearest earlier stage's environment with something applied, the newest first. The first stage holding the stack is compared with nothing, siblings are never compared with each other, and the next stage after the last one holding the stack is drawn too, where Promote would create it.

```text
qa       Live r9 · Raise replicas · app-infra v4
staging  Live r6 · Raise replicas · app-infra v4    Same template as qa
prod     Live r14 · Pin version · app-infra v3      2 revisions behind staging
```

- **What runs**: the live revision and the chart version it runs. A stage with nothing applied reads **Not Applied**.
- **How far from the environment it is compared with**, read from the templates' digests, never from timestamps: **Same template as qa**, **2 revisions behind staging** (Promote brings it here), **Ahead of qa** (the stage before moved away from this template), or **Differs from staging** (it was changed on its own). A roll back is never mistaken for being behind.
- **Update Available** when the chart has a newer version than the stage runs, and **customised** when the stack's files differ from the chart version it is on.
- **What is moving in**, live: a run applying, a request waiting for approval ("Promoting staging r6 · Applying").

The step between two stages offers **Promote**. A stage you can't read says only that. `planton infra stack status <name>` prints the same path, and agents read it with `get_infra_stack_delivery_path`.

## Promote

Promote carries one environment's live template to another stage of its Promotion Path, with its `values.yaml` and its chart version; the target keeps every value it set.

```bash
# To the next stage on staging's path
planton infra stack promote app-infra --from staging -m "Raise replicas"

# Name the target, or pick one sibling of a stage that holds several
planton infra stack promote app-infra --from staging --to prod -m "Raise replicas"
```

Left out, `--to` is the next stage on the source's path: its one environment, or, when that stage holds siblings, the one you choose at a terminal. Each sibling is promoted on its own, as its own plan and run.

- **A value travels only when you bring it**: **Bring This Value** in the dialog, `--bring <name>` on the CLI. A param the travelling `values.yaml` newly requires asks for a value (`--set name=value`); a value for a param it no longer declares is shown as dropped.
- **What the source keeps, the target keeps.** A resource staging removed from its templates but kept running is offered as Keep in prod too ("kept in staging r6"), never silently destroyed. A Destroy, Adopt or Re-apply chosen in one environment never carries to another.
- **A first promotion** into an environment with no stack of that name creates it there, with every value from the source. Promotion runs in either direction along the path, so "make qa like prod" works.
- **A promotion stays on its path.** Across two paths, or into or out of an environment on no path, it is refused, naming both places and the apply that reaches the environment directly: "A promotion stays on its Promotion Path: 'prod-eu' is on the homes path and 'zm-staging' is on the mortgage path. ..."
- **The message** defaults to the source revision's, and the revision, its run, the activity feed and notifications read **"Promoted from staging r6"**.
- **A template that names its source environment** is a warning on the plan, never a block: "templates/db.yaml names qa (line 12). It reads the same in prod. Use {{ values.env }}."
- **Into a protected environment** a promotion is a request, and the approver reads "Promoted from staging r6" above the exact plan.
- **In a Promotion Only stage**, a template edit that did not come by promotion is cautioned on its plan and the path reads "Differs from staging"; it is never refused, and protection still decides who may change the environment.

## Roll Back

Roll Back plans an earlier revision's declaration — its templates, values and chart version — fresh against today's cloud, as a new revision ("Roll back to r8"). From the **Revisions** tab, **Roll Back to r9**; in a terminal:

```bash
planton infra stack rollback app-infra --env qa                  # the default target
planton infra stack rollback app-infra --env qa --to-revision 8
```

The default target is the live revision when the latest one failed, otherwise the newest applied revision before the live one; the target must have applied. Before you apply, know what a roll back does not do:

- **Module versions stay current.** It restores the declaration, not the modules that applied it.
- **Data destroyed since comes back empty.** A database destroyed after that revision is created again, not restored.
- **The old choices are not carried.** A resource kept on its provider since then plans Adopt (`--adopt Kind/name`), and one another stack now owns plans Blocked, naming the owner.

A roll back is refused while a request waits or a run is in flight on the stack.

## Upgrade

A stack installed from a catalog chart can move to the chart's newest version, keeping its values: **Update Available · Upgrade** on the stack's Overview, or

```bash
planton infra stack upgrade app-infra --env qa
```

Each file is compared three ways — the version the stack is on, the stack's own file, and the new version's. A file you never changed takes the new version's; a file only you changed stays yours; a file both changed is a conflict you settle with **Keep Mine** or **Take v5** (`--keep-mine <file>`, `--take-chart <file>`), and the plan cannot apply until each is chosen. `values.yaml` is compared the same way. A file the new version deletes is removed when you take it, and its resources plan Destroy. A value the new version no longer declares reads "Dropped: v5 no longer declares it". When an earlier stage already runs the new version, the later stage reads "qa runs v5 · Promote from qa": promote the template qa tested instead.

## Compare

Compare puts two environments, or two revisions, side by side: values first, each with where it comes from, then how far apart the templates are, the resources only one side declares, and the files that differ. **Make prod Like qa** opens Promote.

```bash
planton infra stack diff app-infra --env prod --against qa
planton infra stack diff app-infra --env qa --revision 10 --against-revision 9
```

`diff` exits `0` when the two are the same and `1` when they differ.

### What Every Plan Adds

Every plan — an edit, a promotion, a roll back or an upgrade — shows above its rows where the declaration came from, and beside them:

- **where each value comes from**: set here, the chart's default, brought from another environment, or dropped;
- **What Else This Affects**: everything that reads a resource the plan updates, replaces, destroys or may change — another stack, a service, a resource deployed on its own ("payments-infra in prod reads orders-db", "checkout-api in prod reads orders-db"). The readers are not blocked, but a destroy of something one of them reads is: that row reads Blocked, and Keep applies the plan around it;
- **what it costs and opens up**: the monthly cost before and after, the permissions it adds, the security settings it changes and anything over budget (see [What a Plan Costs and Opens Up](#what-a-plan-costs-and-opens-up));
- **the template files** it changes and the **chart version** it moves between;
- **warnings**, which block nothing.

Promote, Roll Back and Upgrade take what `planton infra stack apply` takes — `--plan-digest`, `--yes`, `--destroy`, `--keep`, `--adopt` — plus `--dry-run`, `--detailed-exitcode` (exit `4` when applying would change something) and `-o json|markdown|yaml`. In CI, post the plan to the pull request, then apply only that plan:

```bash
planton infra stack promote app-infra --from staging --to prod --dry-run -o markdown
planton infra stack promote app-infra --from staging --to prod -m "Raise replicas" --plan-digest <digest>
```

## Infra Stack Lifecycle

### Plan and Apply

Plan a stack from a chart, a checked-out folder, or the declaration it already holds:

```bash
planton infra stack plan my-stack ./my-chart --env <env> -f values-prod.yaml
planton infra stack apply my-stack ./my-chart --env <env> -f values-prod.yaml -m "why"
```

A values file lists only the values it sets; every other value keeps the chart's default (see [how values resolve](/docs/infrastructure/infra-charts)). With no chart, the plan is of the stack's own declaration — what moved underneath it — and `--set name=value` alone changes one value. `planton chart install my-stack ./my-chart --env <env>` takes the same path and creates the stack when the name is new; `--dry-run` stops at the plan.

At a terminal, the apply shows the plan, says what it destroys and replaces, and asks once. Without a terminal (CI, a script, an agent), `--plan-digest` from a plan a person reviewed applies exactly that plan, and `--yes` applies the fresh plan; either consents to every Destroy and Replace it shows, and the apply prints what it destroyed and replaced. `--keep Kind/name` keeps a removed resource and `--adopt Kind/name` takes one over; `--destroy Kind/name` only turns a Keep carried from an earlier revision back into a Destroy. With none of them and no terminal, the apply refuses, says what the plan destroys and replaces, and prints the same command with `--yes` added.

`planton infra stack plan` exits `0` when planned, `1` when the declaration does not build, a declared resource is blocked or a run is in flight, and `2` when the plan could not be made; `--detailed-exitcode` adds `4` for a plan that would change something. `-o markdown` prints the plan for a pull request.

In the console, **Review Plan** opens the same plan — from the stack page, the Configuration tab, the create form, or Studio — and **Apply** runs it.

### Undeploy

Destroy everything the stack owns without deleting the stack record. The stack's declaration and revisions are kept, and applying its plan again brings it back — useful for temporarily tearing down infrastructure to save costs:

```bash
planton infra stack undeploy <stack-name-or-id> --env <env>
```

An undeploy covers every Infra Component the stack created, including components its templates no longer declare. It reads each resource from its stored record, never from the templates, and destroys dependents first: a workload goes before the cluster it runs on. Each record is settled the moment its own destroy ends:

- A resource the stack holds as **Unmanaged**, or one marked `planton.dev/keep-on-provider: "true"`, is released from the stack and stays on its provider. An undeploy never destroys it.
- A resource whose record is already gone reads **Already Gone**.
- A resource another stack or service now owns is left in place for its owner.

An undeploy is refused before anything starts when:

- any resource the stack owns carries `planton.dev/prevent-destroy: "true"` (see [Protecting a Resource from Teardown](#protecting-a-resource-from-teardown));
- anything else still reads one of its resources — another stack, a service, or a resource deployed on its own, through a reference to it or with workloads placed in its namespace. The refusal names the reader and how to let go of the reference;
- the stack owns no Infra Components: it "owns no infra components, so there is nothing to undeploy".

In a protected environment an undeploy is a request an approver decides, as an apply is: nothing is destroyed until then. The CLI says what goes and asks before undeploying or purging; `--yes` answers ahead, and without a terminal the command refuses and prints itself with `--yes` added. Applying the stack again later plans each resource anew: any resource the undeploy released plans Adopt.

### Purge

Undeploy the stack, then delete its record — the record is deleted only if the undeploy succeeded:

```bash
planton infra stack purge <stack-name-or-id> --env <env>
```

A purge is refused whenever an undeploy would be, with one difference: a purge of a stack that owns nothing simply deletes it.

### Delete

Deleting an Infra Stack never destroys infrastructure, so it is refused while the stack still owns Infra Components — undeploy it first, or purge it to do both. A delete is also refused while a run of the stack is in flight.

## Protecting a Resource from Teardown

Annotate a resource with `planton.dev/prevent-destroy: "true"` and it is never torn down through the platform:

```yaml
metadata:
  name: prod-db
  annotations:
    planton.dev/prevent-destroy: "true"
```

While any resource the stack owns carries the annotation, the stack's undeploy and purge are refused before anything starts, naming the resource. A person's destroy or purge of that Infra Component is refused too, as is creating, rerunning or resuming a destroy job for it. Adding or removing the annotation changes no deployment: no job runs for that change alone.

To lift the protection, remove the annotation from the stack's templates and apply the stack (the plan reads Update Record: nothing in the cloud changes). A plan that removes a protected resource reads it Blocked from destroying — choose Keep to apply around it. For a resource the stack no longer declares, remove the annotation from the Infra Component itself. Deleting a record alone is not a destroy, so the annotation does not stand in its way.

## Keeping a Resource on Its Provider

Annotate a resource with `planton.dev/keep-on-provider: "true"` and its stack never destroys it: whenever the stack lets go of it — removed from the templates, undeployed, purged — it is released, still running, and the plan reads Keep with the choice locked. A released resource can be adopted by another stack or taken over by a service. This is how a resource moves from an Infra Stack to a Service: remove it from the templates, keep it running on its provider, and the service's next deploy takes it over.

A kept resource keeps what it depends on. While a deployment stays, the namespace it lives in — and anything else it references — is never destroyed: the namespace's plan row says what depends on it, with **Keep This Too** in the console and the exact `--keep` flags on the CLI, and an undeploy is refused before anything is torn down, naming each kept resource and what it depends on.

## What a Run Applies

A stack's run applies only the resources whose inputs moved since their last successful deploy — the same comparison its plan shows. Before a resource runs, Planton compares what it would deploy — its settings, the values it reads from other resources, its secrets' current versions and its variables, the account it deploys into, the module it runs, and whether anything it sits on was rebuilt — with what its last successful deploy applied. An exact match runs no job: the resource reads **Unchanged since r11**, naming the revision whose run last applied it, linked to that run, and hands that job's outputs to the resources that depend on it. The job's id stays one click away for its logs (`Unchanged since r11 (ij_…)` in a terminal). Change one value and that resource runs, along with any dependent whose inputs it changes. A module is compared by the digest its release published over every artifact it ships, so a new Planton release runs only the resources whose module that release changed. When Planton cannot know exactly what a resource would deploy — a module that follows a branch, a secret whose version cannot be read, a connection that signs in through the runner's own machine — the resource Always Runs.

A change made outside Planton, such as an edit in the cloud console, moves nothing Planton deploys, so a plan reads that resource Unchanged until a drift check finds it (below).

More outcomes appear on a run's resources:

- **Unmanaged** — a resource whose lifecycle events are disabled (`planton.dev/infra-job.disable-on-lifecycle-events: "true"`, see [Flow Control](/docs/infrastructure/flow-control)). The stack records it and never applies or destroys it; removing it from the templates plans Keep, and an undeploy releases it.
- **Kept** — a removed resource the stack let go of; no job ran, and it keeps running.
- **Owned Elsewhere** — a resource another live Infra Stack or Service owns. A teardown leaves it in place for its owner, and a plan that would change it reads Blocked, naming the owner to change it through.

## Changed Outside Planton: Drift

Choose **Check for Drift** on a stack's resources, or run `planton infra stack drift check <stack-name> --env <env>`, and Planton compares every resource the stack holds with what is really in the cloud, setting by setting. A drift check writes nothing and changes nothing, so it never waits for an approver, even in a protected environment. Each resource reads **In Sync**, **Drifted** (with the number of changes), or **Couldn't Check** (with the reason). One resource can be checked on its own page, or with `planton infra component drift check <infra-component-id>`.

**View Changes** lists each change under **Applied by Planton** and **In the Cloud Now** — `instance_class · db.t3.medium → db.t3.large` — with a secret shown only as "Hidden value changed" (an earlier value of a rotated secret too, unless its backend has deleted that version) and a resource removed by hand as "Deleted Outside Planton".

Once a check finds drift, the next plan of the stack includes that resource as Update, with the reason "Drifted: 3 changes made outside Planton", and applying it puts the declaration back. You choose what happens:

- **Reconcile** — a plan of the declaration the stack already holds, titled "Reconcile drift on orders-db and cache", reviewed and applied like any plan (in a protected environment, a request for an approver). A stack is never applied in part, so a resource that also changes for another reason stays in the plan with a warning naming it. From a terminal: `planton infra stack reconcile <stack-name> --env <env>`.
- **Edit in Template** — keep the change made in the cloud: open the template file that declares the resource, carry the value over, and apply. The next plan has nothing left to put back.

A successful apply clears the resource's Drifted mark. The person who ran a stack's check is notified when it finds drift or can't check something, and every check a person runs appears in Activity.

## Using the CLI

```bash
# Plan: what applying would do, resource by resource (writes nothing)
planton infra stack plan my-stack ./chart-dir --env <env> -f values.yaml

# Apply exactly the plan shown, as a numbered revision, and follow its run
planton infra stack apply my-stack ./chart-dir --env <env> -f values.yaml -m "why"

# Install a chart as a stack: the same plan, review and apply
planton chart install my-stack ./chart-dir --env <env> -f values.yaml -m "why"

# Install from the catalog: the stack records the chart and its version
planton chart install my-stack app-infra --env <env> -f values.yaml -m "why"

# A stack's revisions, newest first, each with where it came from
planton infra stack history <stack-name> --env <env>

# Where the stack stands in each environment along its Promotion Path
planton infra stack status <stack-name>

# Carry staging's live template to the next stage on its path (or name it with --to); the target keeps its values
planton infra stack promote <stack-name> --from staging -m "why"
planton infra stack promote <stack-name> --from staging --to prod -m "why"

# Return to an earlier revision, planned fresh
planton infra stack rollback <stack-name> --env <env> --to-revision 8

# Move to the chart's newest version, keeping your values
planton infra stack upgrade <stack-name> --env <env>

# Compare two environments, or two revisions
planton infra stack diff <stack-name> --env prod --against qa

# Check every resource for changes made outside Planton, and read the result
planton infra stack drift check <stack-name> --env <env>
planton infra stack drift show <stack-name> --env <env>

# Put drifted resources back: a plan you review, then apply
planton infra stack reconcile <stack-name> --env <env>

# List runs for a stack, and its last run
planton infra stack infra-pipelines <stack-id>
planton infra stack last-pipeline <stack-id>

# Undeploy (destroy resources, keep the stack and its revisions)
planton infra stack undeploy <stack-name-or-id> --env <env>

# Purge (destroy resources and delete the stack)
planton infra stack purge <stack-name-or-id> --env <env>

# Get stack details: its declaration and live revision
planton infra stack get <stack-name-or-id> --env <env>

# Check a stack out as a working copy: templates, values.yaml, values.<env>.yaml
planton infra stack checkout <stack-name-or-id> --env <env>
```

## When to Use Infra Stacks vs. Direct Infra Components

| Scenario | Recommended Approach |
|----------|---------------------|
| Complete environment (dev, staging, production) | Infra Stack — coordinated deployment, dependency handling, one recorded configuration |
| Single resource needed quickly | Direct Infra Component — no orchestration overhead |
| Reproducible deployments | Infra Stack — configuration captured as a persistent record |
| Quick experiments | Direct Infra Component — fast iteration, easy to delete |
| Production infrastructure | Infra Stack — reviewed plans, revisions, requests an approver decides on the exact plan, protected teardown |
| Adding to existing infrastructure | Direct Infra Component — targeted change without touching other resources |

## Related Documentation

- [Infra Charts](/docs/infrastructure/infra-charts) — The templates that Infra Stacks instantiate
- [Infra Pipelines](/docs/infrastructure/infra-pipelines) — How Infra Stack deployments are orchestrated
- [Infra Components](/docs/infrastructure/infra-components) — The deployed instances that Infra Stacks own and manage
- [Infra Jobs](/docs/infrastructure/infra-jobs) — The atomic execution units within pipelines
- [Flow Control](/docs/infrastructure/flow-control) — What a resource's annotations and a protected environment decide about how its jobs run
