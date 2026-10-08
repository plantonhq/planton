---
title: "How to Configure Branch Deployments and Tag Releases"
date: "2026-04-02"
author:
  - name: "Planton Team"
    title: "Platform Engineering"
    bio: "Helping teams deploy infrastructure and services without the DevOps bottleneck"
tags:
  - "service-hub"
  - "branches"
  - "deployment"
  - "environments"
  - "pipeline"
category: "service-hub"
excerpt: "Control which Git events build, which environments a run walks, how protected environments gate a release, and how tags and pull requests join in."
---

# How to Configure Branch Deployments and Tag Releases

This tutorial shows you how to control the relationship between Git activity and deployments in Planton. You will learn which pushes build, how a run walks your environments in promotion order, how a protected environment pauses a run for approval, how to build (or deploy) pull requests and tags, how to bind a standing branch to exactly one environment, and how to start a run by hand.

If you completed [How to Deploy Your First Service with Zero-Config CI/CD](/tutorials/how-to-deploy-your-first-service-with-zero-config-cicd), you have a Service that builds on pushes to `main` and deploys the overlays your repository carries. This tutorial builds on that foundation.

> **Note**: The Planton web console offers the same settings on the service's page, and the Planton Assistant makes every edit below when asked in plain words ("also build pull requests", "deploy tags that look like v*"). This tutorial uses the YAML so every step is exact and reproducible.

## What You Will Learn

- Where every trigger rule lives: `spec.build.triggers`
- How a run walks your environments, and why the order and the gates are the environments' properties, not the service's
- How to build pull requests, and how to give each one a preview environment
- How to set up tag-based releases with glob patterns
- How to bind a branch to one environment (`spec.deploy.branchDeployments`)
- How to start, follow, re-run, and cancel a run from the CLI

## Prerequisites

- [ ] A working Service deployed through Planton (see [How to Deploy Your First Service with Zero-Config CI/CD](/tutorials/how-to-deploy-your-first-service-with-zero-config-cicd))
- [ ] At least two environments in your Planton organization (e.g., `dev` and `production`)
- [ ] The `planton` CLI installed and authenticated

## How Triggers and Deployments Relate

Two things decide what happens on a push. **Triggers** (`spec.build.triggers`) decide whether the push BUILDS: which branches, which paths, whether pull requests and tags count. **Your environments** decide where a successful build DEPLOYS: a run walks every environment the service declares, in the organization's promotion order, and pauses wherever an environment is protected. The service does not list environments in a second place — for a git-maintained service the overlays under `_kustomize/overlays/<env>/` ARE the environment set; for a manually configured service, the entries under `spec.deploy.environments` are.

## Step 1: Deploy to a Second Environment

Your service deploys to `dev`. To add `production`, add its overlay to your repository:

**`_kustomize/overlays/production/kustomization.yaml`**

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - path: service.yaml
```

**`_kustomize/overlays/production/service.yaml`**

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesDeployment
metadata:
  name: my-service
  env: production
spec:
  availability:
    minReplicas: 3
  container:
    app:
      resources:
        requests:
          cpu: 250m
          memory: 512Mi
        limits:
          cpu: 2000m
          memory: 2Gi
```

Commit and push. That push is itself a run: the platform reads the new overlay, records `production` as one of the service's environments, and from now on every run walks `dev` and then `production`, in the promotion order your organization defines for those environments. If `dev` fails, `production` is never attempted; completed environments stay deployed.

## Step 2: Gate Production with an Approval

Whether a run may enter an environment without a human is the ENVIRONMENT's property — set it once on the environment (protected), and every service that deploys there inherits the gate. When a run reaches a protected environment it pauses with status `awaiting_approval` and stays there until someone decides.

To see what is waiting on you:

```bash
planton service pipelines --awaiting-my-approval
planton service last-pipeline my-service        # the run's record; the paused environment carries requiresManualGate: true
planton service follow <run-id>                 # live status
```

To approve or reject:

```bash
planton service resolve-env-manual-gate <run-id> production yes
planton service resolve-env-manual-gate <run-id> production no
```

Two rules hold on every gate: the person who set the run in motion cannot approve it into a protected environment (separation of duties, enforced by the server), and the decision, the decider, and their note are recorded on the run — the durable "approved by" answer.

## Step 3: Build Pull Requests

By default a pull request does nothing. To build every pull request whose target branch is one of your trigger branches — the image is built and pushed tagged with the commit sha, nothing is deployed — turn on the pull-request trigger:

```yaml
spec:
  build:
    triggers:
      pullRequests:
        build: true
```

```bash
planton apply -f service.yaml
```

The run's deploy stage is skipped with the reason written on the record ("Pull request runs build only -- enable build.triggers.pullRequests.deploy ..."), and the build's verdict shows on the pull request as a check named after the service.

## Step 4: Give Each Pull Request a Preview Environment

To deploy every pull request into its own preview environment (`<service>-pr-<number>`), turn on preview deploys — `deploy` implies `build`:

```yaml
spec:
  build:
    triggers:
      pullRequests:
        deploy: true
        previewTtlHours: 72     # optional; the default. Every push refreshes the clock.
```

A preview is torn down automatically when the pull request closes, or when it goes `previewTtlHours` without a new push (the default is 72 hours; the maximum is 720). Preview-specific deltas — a smaller replica count, a preview hostname — live in your repository under `_kustomize/previews/<env>/` and are rendered per run. The full preview model is in the [preview environments documentation](/docs/ci-cd/deployment-environments).

## Step 5: Configure Tag-Based Releases

A tag is a release. A tag run builds the image tagged with the TAG NAME (a branch run tags with the commit sha) and, when tag deploys are enabled, rides the full promotion walk — gates and protection included — exactly like a default-branch push.

Build tags without deploying them:

```yaml
spec:
  build:
    triggers:
      tags:
        build: true
        patterns:
          - "v*"
```

Build and deploy tags (`deploy` implies `build`):

```yaml
spec:
  build:
    triggers:
      tags:
        deploy: true
        patterns:
          - "v*"
          - "release-*"
```

`patterns` are globs: `v*` matches `v1.0.0`, `v2.1.3-beta`, `v0.0.1-rc.1`; `release-*` matches `release-2026-04-02`. A tag matching ANY pattern triggers; an empty list means every tag triggers once tag builds or deploys are on. A build-only tag run skips its deploy stage with the reason on the record ("Tag runs build only -- enable build.triggers.tags.deploy ...").

## Step 6: Bind a Branch to One Environment

Some teams promote by merging between standing branches instead of walking one branch through every environment. Map a branch to exactly one environment:

```yaml
spec:
  deploy:
    branchDeployments:
      - branch: staging
        env: staging
```

A push to `staging` builds and deploys into `staging` only — it never walks the promotion order — and inherits that environment's gate and protection. A branch may be a trigger branch OR a mapped branch, never both; the platform refuses the ambiguity when you apply. Pushes to the trigger branches keep the promotion walk.

## Step 7: Start, Re-run, and Cancel Runs by Hand

```bash
planton service run my-service --branch main                    # build the head of main
planton service run my-service --branch main --commit a1b2c3d4  # build a specific commit
planton service run my-service --branch main --deploy-env dev   # deploy into exactly one environment, no walk
planton service run my-service --branch staging                 # a mapped branch deploys to its environment, as its push would
planton service rerun <run-id>                                  # byte-identical: the same compiled pipeline, the same params
planton service cancel <run-id>
planton service runs my-service                                 # everything that ran, newest first
```

A manual run goes through the same path as a push: the same compile, the same build, and the same destination. A trigger branch walks the promotion order, and a mapped branch deploys only to its environment; `--deploy-env` names one environment instead. A service with no trigger branch has no walk for a commit to ride, so `--commit` there needs `--branch` (deploying where that branch is mapped) or `--deploy-env`, and is refused without one. A rerun replays the run's stamped pipeline definition without recompiling, so it is the right verb after a fix outside the code (a credential, a runner); after a code fix, push a new commit.

## Common Patterns and Tips

### Build-only services

A service that should build and publish but not deploy — a shared base image, a service deployed by another system:

```yaml
spec:
  deploy:
    disableDeployments: true
```

Every run builds and pushes the image; the deploy stage is skipped with the reason on the record.

### Pausing all automation

```yaml
spec:
  build:
    triggers:
      disabled: true
```

Pushes, tags, pull requests, and manual runs all stop until you remove it.

### Trigger branches and trigger paths together

`spec.build.triggers.branches` (empty means the repository's default branch) and `spec.build.triggers.paths` are both conditions: the push must land on a trigger branch AND change a file under the project root or a trigger path. See [Monorepo Support](/docs/ci-cd/monorepo-support).

### A protected environment and the initiator

If your own push is waiting at `production` and you are the only administrator, the platform will still refuse your approval — that is the point of the gate. Add a second approver to the organization, or leave the environment unprotected.

## What to Do Next

- **Write your own pipeline** for custom build steps like SonarQube analysis, security scanning, or multi-stage builds. See [Self-Managed Pipelines](/docs/ci-cd/self-managed-pipelines).
- **Understand the run record** — what a skipped stage says and how to read a failure. See [Pipelines](/docs/ci-cd/pipelines).
