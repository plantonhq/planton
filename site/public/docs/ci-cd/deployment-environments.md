---
title: "Deployment Environments"
description: "Which environments a service deploys to, how a push walks them along your Promotion Paths, and how a branch can be pointed at exactly one."
icon: deployment
order: 50
tags:
  - Deployment
  - Environments
  - Promotion Paths
  - Service Hub
---

# Deployment Environments

A service deploys to the environments it declares configuration for, and a push walks them along your organization's **Promotion Paths**: the order, declared once for the whole organization, in which work moves through its environments. That is the right behavior for most services: a push to the trigger branch builds once and walks dev, then staging, then production, stopping at any protected environment until someone approves, or where the path says the rest is reached only by promotion. Sometimes you need something narrower — a canary service that should only reach staging, a team that keeps one long-lived branch per environment and promotes by merging, a new service that should start in dev alone and grow. This page covers how the set of environments is decided, how Promotion Paths order them, and how a branch can be aimed at one of them.

<!-- VIDEO: Deployment environments walkthrough
  Source: Cloudflare Stream
  Alt: Video walkthrough showing a service's environments on the Configuration tab and a branch mapped to one environment
-->

## How It Works

Every service carries one configuration section with **one entry per environment it deploys to**. Each entry holds the environment's slug and the full manifests the service deploys there. When a pipeline reaches its deploy stage, it creates one deployment task per declared environment and runs them along the Promotion Paths those environments sit on (see [Promotion Paths](#promotion-paths-the-order-a-push-walks)). An environment with no entry is not deployed to — there is no separate filter to maintain.

Who writes those entries depends on the service:

- **Platform-authored services** declare their environments on the Configuration tab (or in `service.yaml`). Add an environment and its resources, and the next run deploys there; remove its configuration and the next run does not.
- **Git-maintained services** (a `_kustomize` tree in the repository) get their entries from the repository: **the overlay set defines the environment set**. An `overlays/staging/` directory means the service deploys to staging; delete the overlay and the service stops deploying there on the next push. The platform syncs the rendered overlays onto the service record, and the record shows which branch and commit each environment's configuration came from.

Two sibling trees can sit beside the overlays without joining the environment set: `dev/<flavor>/` serves local development only (never deployed), and `previews/<env>/` holds the deltas a pull-request preview of that environment deploys with (rendered per run, never written onto the record).

<!-- SCREENSHOT: A service's environments on the Configuration tab
  Page: /{org}/service/{slug}?view=configuration (Configuration tab)
  Action: A git-maintained service with dev, staging, and prod entries, each panel headed by what is live there and listing the declared resources beneath, with the sync provenance (branch, commit, when) on each
  Focus: The three environment panels, in the order a push walks them
  Alt: Configuration tab showing one panel per environment the service deploys to
-->

## Choosing Which Environments a Service Deploys To

### Web Console

Open the service and go to **Configuration**. Each environment the service deploys to is one panel: the environment's name and what is live there at the top, the declared resources beneath. On a platform-authored service, **Add Environment** declares a new one and **Remove Configuration** on a panel stops deploying there; each declared resource has its own **Edit**. On a git-maintained service the panels are read-only and each carries **Open in Repository** — the overlay is where the change is made, and the next push brings it back.

### CLI

For a platform-authored service, the environment entries live under `spec.deploy.environments` in your service YAML:

```yaml
spec:
  deploy:
    environments:
      - env: dev
        resources:
          - apiVersion: kubernetes.planton.dev/v1alpha1
            kind: KubernetesDeployment
            # ...
      - env: staging
        resources:
          # ...
```

```bash
planton apply -f service.yaml
```

For a git-maintained service, edit the overlays and push — an apply cannot set or clear the synced entries, exactly as it cannot write a record's status.

## Promotion Paths: the Order a Push Walks

The order your environments come in is not a setting of any service. Your organization declares it once, as its **Promotion Paths**, and every service and every Infra Stack in the organization reads the same answer.

- **A Promotion Path** is one product line's way through your environments: a name and its stages, first to last. A retail team might run `dev → staging → prod`; a second product line in the same organization runs `zm-dev → zm-staging → zm-prod` on a path of its own.
- **A stage** holds one environment, or several side by side. Environments in one stage are **siblings**: each comes after every environment of the stages before it, and there is no order among them. `dev → {demo, prod}` is two stages, and demo and prod are both next after dev.
- **An environment sits on at most one stage of one path.** Paths run in the order your organization lists them.
- **Promotion Only.** A path can mark a stage Promotion Only: that stage and every stage after it change only by promotion (or by deploying a build there directly). On `dev → staging → prod` with Promotion Only from stage 3, pushes deploy dev and staging, and prod waits for someone to promote.

A service never names a path. It follows the paths its declared environments sit on, so the same service can span two product lines' paths without saying so.

### What a Push Walks

A push to a trigger branch, a tag release, and a manual run by commit each walk every path the service's environments sit on: path by path in your organization's order, stage by stage, a stage's environments one after another by name, stopping on each path where Promotion Only starts. A failure stops the whole run, so on a service that spans two paths, a failure on the first means the second is never attempted.

The run still reads every environment the service declares and records why it did not deploy the ones it passed:

- **An environment on no path is skipped, never guessed into an order.** Planton has no way to know whether an unplaced `qa` belongs before or after `prod`, so a push leaves it alone and says so: "Not on a Promotion Path: qa is on no path, so pushes don't deploy it — add it to a path for pushes to reach it, or deploy this build to qa". The run keeps the environment's configuration with the build, so deploying that build there later works.
- **An environment a mapped branch deploys** is left to its branch: "Deployed by its branch: pushes to release deploy prod, so this run leaves it alone". The walk goes on past it.
- **An environment in a Promotion Only stage** reads **Awaiting Promotion**.

A run left with nothing to deploy ends **Skipped**, never Succeeded, and says why: "Nothing to deploy: qa and staging are on no Promotion Path, and pushes to hotfix deploy prod — the build is kept; deploy it to an environment to run it". An organization with no paths at all walks no environment on a push; branch-mapped deploys, direct deploys and previews never read an order, so they work the same.

Environments that hold shared infrastructure for every product line — a `preprod` with shared databases, a `shared` networking environment — usually sit on no path. They keep serving every path through references; they just never receive a push.

### Setting Your Promotion Paths

#### Web Console

Go to **Environments → Promotion**. The **Promotion Paths & Protection** page shows one panel per path: its stages left to right, siblings side by side, Promotion Only from the stage where it starts, and each environment's protection. Each environment has **Move To**, each stage moves up and down, and **New Path** and **Delete Path** add and remove whole paths. Environments on no path are listed beneath under **Not on a Path**, each placed with its **Move To** menu. Before any change that makes pushes start or stop deploying an environment, the confirm shows the path Now beside After and says what changes ("Pushes will then stop deploying zm-staging."), and it names any branch mapped to an environment the change makes reachable only by promotion.

**New Environment** asks where the new environment goes: a new last stage of one of your paths (a path named after your organization when it has none yet), or on no path.

<!-- SCREENSHOT: Promotion Paths & Protection
  Page: /orgs/{org}/environments/promotion
  Action: An organization with two paths, one of them dev → staging → {demo, prod} with Promotion Only from stage 3, and one environment on no path
  Focus: The path panels, the sibling stage, and the Not on a Path section
  Alt: Promotion Paths page showing two paths with their stages, a sibling stage, and an environment waiting to be placed
-->

#### CLI

```bash
# Every path on one line, in the order pushes walk them; the environments on no path close the list
planton promotion-path list

# Set one path whole: each --stage is a stage, commas make siblings
planton promotion-path set retail --stage dev --stage staging --stage demo,prod --promotion-only-from 3

# Delete a path; its environments go on no path
planton promotion-path delete mortgage

# Where one environment sits
planton env get prod
```

`set` replaces the path of that name, or adds it after your last path, and asks first when the change makes pushes start or stop deploying an environment (`--yes` answers ahead of time). `delete` always asks. `planton env create` places the new environment as it creates it: at a terminal it asks where it goes, `--path <name>` places it as a new last stage of that path (with `--stage <n>` to join stage n beside its environments), and `--no-path` leaves it on none. In a script it lands on no path and prints the command to place it.

The whole record also applies as a document:

```yaml
apiVersion: tenancy.planton.ai/v1alpha1
kind: PromotionPathSet
metadata:
  org: acme
spec:
  paths:
    - name: retail
      stages:
        - environments: [dev]
        - environments: [staging]
        - environments: [demo, prod]
      reachedByPromotionFromStage: 3
```

```bash
planton apply -f promotion-paths.yaml
```

`reachedByPromotionFromStage` is Promotion Only: the first stage, counted from 1, that changes only by promotion (0, or left out, when pushes deploy every stage). Coding agents use the same record through `get_promotion_path_set`, `set_promotion_path`, `delete_promotion_path` and `apply_promotion_path_set`.

#### Who May Change Them

Anyone in the organization can read the paths; a developer needs them to know where a build goes next. A change needs update on the organization's Promotion Paths (Infrastructure Admin and the organization owner hold it) and update on every environment whose place it changes — so moving a protected `prod` needs an explicit grant on `prod`. Every save is made from the version you read: if someone changed the paths in between, nothing is saved, and the answer is "Promotion Paths changed since you opened them; reload them and make your change again". Deleting an environment takes it off its path, and a stage or path left empty disappears.

## Branch-Based Targeting

By default a push to a trigger branch walks every declared environment along its Promotion Paths. Teams that keep a standing branch per environment can instead map a branch to exactly one environment. A push to a mapped branch builds and deploys into **that environment alone**, wherever it sits — it never walks the paths — and inherits the environment's gates and protection. The one exception is an environment in a Promotion Only stage: a mapping there deploys nothing, because that stage changes only by promotion. Promotion then happens the way those teams already work: by merging between branches.

A branch is either a trigger branch (walks everything) or a mapped branch (deploys to one environment), never both; a configuration that tries to make it both is refused when it is saved, naming the ambiguity. Mappings name a branch exactly — no patterns — because a mapping is a deliberate, auditable statement about a standing branch.

### Web Console

On the Configuration tab, the **Deploy Settings** panel's **Branch Deployments** row lists the mappings. **Edit** opens a dialog of branch → environment rows over the organization's environments; the environment must be one the service declares configuration for. If a branch you map is also in the trigger branches, the dialog says so above the field and offers a door to the Build Triggers dial to resolve it.

<!-- SCREENSHOT: Branch Deployments
  Page: /{org}/service/{slug}?view=configuration (Configuration tab)
  Action: The Deploy Settings panel with a Branch Deployments row reading "release-hotfix → prod" and its Edit dialog open
  Focus: The Branch Deployments row and the dialog's branch → environment rows
  Alt: Branch Deployments row on the Deploy Settings panel with its edit dialog
-->

### CLI

```yaml
spec:
  build:
    triggers:
      branches:
        - main            # walks the paths: dev → staging → prod
  deploy:
    branchDeployments:
      - branch: release-hotfix
        env: prod         # deploys to prod only
```

### Example: One Branch per Environment

```yaml
spec:
  build:
    triggers:
      disableTriggerBranches: true   # no push walks the Promotion Paths
  deploy:
    branchDeployments:
      - branch: dev
        env: dev
      - branch: staging
        env: staging
      - branch: release
        env: prod
```

With this setup:
- Push to `dev` → deploys only to dev
- Push to `staging` → deploys only to staging
- Push to `release` → deploys only to production
- Push to any other branch, `main` included → builds nothing

`disableTriggerBranches` is what keeps the default branch out of it. An empty `branches` list is not the same thing: it means the repository's default branch, so `main` would still build and walk the paths. With no trigger branch the mapped branches still build into their one environment, and tag releases still work if the service turns them on.

For a git-maintained service a mapping is also the **sync authority** for its environment: the mapped branch's push writes that environment's entry from its own tree, so an overlay that exists only on that branch lands on the record marked for its environment.

## Use Cases

### Branch-Based Deployments Without Merge Conflicts

Teams that keep separate branches per environment used to fight over per-environment directories at merge time. With branch mappings, every branch can carry the whole `_kustomize` tree: which overlay actually deploys is decided by the branch the push came from, not by what the tree contains.

### Progressive Rollout

Start a new service with one environment and add the rest as confidence grows — on a platform-authored service, Add Environment on the Configuration tab; on a git-maintained one, add the overlay directory and push. If the environment already sits on a Promotion Path, nothing else changes: the next run walks it at its place on the path. An environment on no path is skipped, with the reason, until you place it.

### Environment-Specific Services

Some services should never deploy everywhere — debug tooling in dev only, internal monitoring in the non-production environments. Declare configuration for exactly those environments and no others; the deploy stage has nothing to do for the rest.

## Where Each Environment Stands: the Delivery Path

A service's **Overview** opens on its **Delivery Path**: its environments along the Promotion Paths they sit on, one row per path (headed by the path's name when there are several), stage by stage, siblings in one stage sharing a column, each with its own step into it. Each tile shows the build that runs there (its commit and the run that built it, "Built by Run #128"), its rollout verdict, its address, and how far it is from the environment it is compared with:

- **3 commits behind dev** — the later environment runs an older commit of the same line.
- **Same build as dev** — both run the very same build; nothing to move.
- **Same commit as dev, rebuilt** — the same commit, built separately.
- **Diverged from dev: 2 ahead, 3 behind** — each side has commits the other lacks.
- **Runs an older build than dev** — the commits could not be compared (a GitLab repository, a comparison that timed out), so the build order answers.
- **Nothing Deployed** — nothing has reached this environment yet.

An environment in a path's first stage is compared with the head of the branch its build came from. A later one is compared with the environment its build was promoted from, when that sits in an earlier stage; otherwise with the nearest earlier stage's environment that holds a build, the newest first. Siblings are never compared with each other, and nothing is compared across paths. Environments a branch deploys into directly read "Driven by release" and are compared with their own branch ("2 commits behind release"); environments in a Promotion Only stage read "Changes only by promotion". Environments on no path follow the paths under **Not on a Promotion Path**, with **Place Them** and no step.

A run moving something into an environment plays on its tile until it lands — "Promoting 'Fix login' from dev · Deploying · Run #129", or **Awaiting Your Approval** with **Review** when the decision is yours. The path is live: a promotion started from the CLI, an agent, or a push appears within a second, with no reload.

**Promote lives on the step between two environments.** The step offers Promote when the earlier environment has a build and the later one is behind it or empty. It still offers Promote, behind a one-line caution, when the later environment is ahead, has diverged, was just rolled back, or the earlier one failed its checks. It stays quiet when both run the same build, or when a run already in flight will carry the build there ("Arrives with Run #130"). The confirm asks exactly what will happen — "Promote 'Fix login' to staging?" — and names where the build comes from and which run built it; when an earlier stage holds more than one build, **Change** lists the others. **Roll Back** lives in the **Deployments** tab, beside the history it restores from.

**Promote stays on its path.** A build moves between stages of the same path, in either direction, and never across paths or into or out of an environment on no path; that is refused with the reason and the way through ("promotion stays on its Promotion Path: 'zm-staging' is on the mortgage path and 'staging' is on the retail path. ..."). To put a build into an environment on another path or on no path, deploy it there. **Each sibling is promoted on its own**: on `dev → {demo, prod}`, promoting into demo and promoting into prod are two independent runs, and a protected one pauses at its own gate.

```bash
# Promote dev's build to the next stage on its path
planton service promote my-service --from dev

# Name the target when you want one sibling in particular
planton service promote my-service --from dev --to prod
```

Left out, `--to` is the next stage on the source's path: its one environment, or, when that stage holds siblings, the one you choose at a terminal. In a script, a stage of siblings is refused, naming the `--to` to add.

<!-- SCREENSHOT: The Delivery Path
  Page: /{org}/service/{slug}
  Action: A service with dev, staging and prod; staging three commits behind dev with Promote on the dev → staging step; prod running the same build as staging
  Focus: The environment tiles along the path and the steps between them
  Alt: Service Overview showing three environment tiles on one Promotion Path with Promote on the step between dev and staging
-->

From a terminal, `planton service status <service>` prints the same path in the same words — each environment's build, rollout, distance, and the runs in flight by number — and `-o json` returns it for scripts and agents.

## How It Appears in Pipelines

A run's page draws the pipeline as a strip: Build, then one tile per environment, path by path and stage by stage, siblings sharing a column. Environments the service does not declare configuration for do not appear, and a push to a mapped branch shows exactly one environment tile. An environment the push left alone stands apart and says why: **Not on a Promotion Path**, with **Open Promotion Paths** as the way to place it, or "Driven by release" for one a mapped branch deploys. When a push stops before an environment that takes builds only by promotion, that environment reads **Awaiting Promotion**, and Promote sits on the strip's step into it — promoting from the run's page keeps you there and turns the environment's panel into the promotion, live.

<!-- SCREENSHOT: Pipeline run page stage strip
  Page: /{org}/service/{slug}/runs/{runId}
  Action: A run from a push to a mapped branch — the strip shows Build and one environment tile
  Focus: The stage strip under the run's header
  Alt: Run page whose stage strip shows Build and a single environment
-->

## Troubleshooting

### An Environment Is Not Deploying

- **Check the Configuration tab** — the environment must have a panel. On a git-maintained service, that means an `overlays/<env>/` directory on the branch that drives the environment.
- **Check Branch Deployments** — if the push came from a mapped branch, only that branch's environment deploys.
- **Check the environment exists in the organization** — an entry names an environment by its slug, and the slug must exist.
- **Check its place on your Promotion Paths** — an environment on no path reads **Not on a Promotion Path** on the run, and no trigger-branch push or tag release deploys it (a branch mapped to it still does). Place it on **Environments → Promotion** or with `planton promotion-path set`. An environment that reads **Awaiting Promotion** is in a Promotion Only stage: promote into it, or move where Promotion Only starts.

### A Push Deploys to One Environment When You Expected the Walk

The branch is mapped under Branch Deployments. Remove the mapping to let pushes on that branch walk the Promotion Paths — or add the branch to the trigger branches after removing the mapping; it cannot be both.

## Preview Environments per Pull Request

Beside the durable environments you configure, the platform mints **preview environments** on its own: when a pull request opens on a service with pull-request previews enabled, a real, short-lived environment named `{service}-pr-{number}` is born from the environment the PR's target branch deploys to, and the changed service deploys into it alone. Rollout verification stamps a working URL for the pull request's own copy of the service.

Three things to know about them:

- **They manage their own lifecycle.** A preview is created by the pull request and destroyed by it — closing the PR (merged or not) tears down its Infra Components and records, and an untouched preview expires on its own (72 hours by default, tunable per service). Deleting a preview environment by hand is refused; closing the pull request is the delete button.
- **They never join a Promotion Path.** A preview cannot be placed on one, so your dev → staging → production walk is unchanged no matter how many previews exist.
- **They are capped.** At most five previews per service exist at once; a pull request beyond the cap still builds, and its deploy explains the skip.

Check any pull request's preview — its phase, verified URL, and rollout verdict — with `planton service previews <service> --pr <n>`.

## Related Documentation

- [Deployment Stage](/docs/ci-cd/deployment-stage) — How manifests are resolved and deployed, including the Kustomize model
- [Deployment Targets](/docs/ci-cd/deployment-targets) — Supported platforms and Git-based vs inline configuration
- [What is a Service?](/docs/ci-cd/what-is-a-service) — Service configuration overview
- [Pipelines](/docs/ci-cd/pipelines) — The pipeline execution model, including pull-request previews
