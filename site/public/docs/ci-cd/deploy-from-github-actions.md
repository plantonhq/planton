---
title: "Deploy from GitHub Actions"
description: "Deploy your service to your own cloud from GitHub Actions — with no backend service anywhere, or connected through your Planton backend. One action serves both, and switching between them is a few lines."
icon: rocket
order: 35
tags: [CI/CD, GitHub Actions, Deploy, Offline]
---

# Deploy from GitHub Actions

GitHub Actions is excellent at building your code — and then most teams hand-roll `gcloud` or `aws` commands to deploy what they built. Planton closes that gap with one published action, `plantonhq/planton/actions/deploy`, that works in two modes:

- **Offline** — no Planton backend anywhere. Your repository carries its own deployment declaration; the runner renders it, verifies everything verifiable in one preflight report, and deploys to your cloud through the open-source engine, keylessly.
- **Connected** — through your Planton backend, with pipelines, approval gates, and rollout verification.

The mode is inferred from the inputs: set `org` and `audience` and you are connected; leave both out and you are offline. This page walks the offline path first — it needs nothing but a cloud account — and shows the one-table switch to connected at the end.

## Why this works with no backend

Your repository's `_kustomize/` tree declares the service's resources per environment — a Cloud Run service and its Redis, an ECS task and its queue — as plain manifests wired together with `valueFrom` references. The CLI derives the deploy order from those references, verifies everything up front (schema, references, state backend reachability, cloud credentials, module availability — one report, before anything is handed to an IaC engine), then deploys each resource through the same open-source modules the platform uses, feeding each resource's outputs into the next one's references. State lives in your own bucket. Credentials come from GitHub's own OIDC exchange with your cloud. Nothing Planton-hosted participates.

## Ten minutes to the first deploy

**1. Declare the deployment in your repository** — a `_kustomize/` tree with a base and one overlay per environment. Leave the workload's container image BLANK: that is the image slot the action fills with the image your job just built.

**2. Give the runner keyless cloud access** with your provider's own official OIDC action — `google-github-actions/auth`, `aws-actions/configure-aws-credentials`, or `azure/login`. No stored cloud keys.

**3. Add the workflow:**

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      id-token: write   # lets the provider's auth action mint an OIDC token
      contents: read
    steps:
      - uses: actions/checkout@v4

      - uses: google-github-actions/auth@v2
        with:
          workload_identity_provider: projects/000000/locations/global/workloadIdentityPools/github/providers/github
          service_account: deployer@my-project.iam.gserviceaccount.com

      # ... build and push your image ...

      - uses: plantonhq/planton/actions/deploy@main
        env:
          # Runners are ephemeral — state must outlive them.
          PLANTON_BACKEND_TYPE: gcs
          PLANTON_BACKEND_BUCKET: my-project-tofu-state
        with:
          environment: prod
          image: ghcr.io/acme/storefront@${{ steps.build.outputs.digest }}
```

**4. Push.** The job log shows the preflight report in a collapsible group — every check line-itemed, every failure a sentence naming the field and the fix — with the verdict outside it. Then the resources deploy in dependency order, and the job's exit code tells the truth.

## What the exit codes mean

- **Refused at preflight** — the report names every problem (an unresolvable reference, an unreachable state bucket, invalid credentials, a missing module) and its fix. Nothing ran; nothing changed anywhere.
- **A resource failed to deploy** — the engine's own error is in the log, verbatim. Completed resources are safe in state: re-running the job re-applies them as no-ops and continues from the failure.
- **Green** — every resource in the tree is live.

## Switching to a Planton backend later

The same action serves the whole journey. When you want pipelines, approval gates, deployment history, and rollout verification, add `org`, `audience`, and `service` to the same step and drop the state env — the backend holds state, and `image` stays exactly as it is. Connected mode needs one trust, made once: [Trusted Workflows](/docs/ci-cd/trusted-workflows) walks it in the console and hands back the whole workflow, filled in. The [action's README](https://github.com/plantonhq/planton/tree/main/actions/deploy) carries the full input table, the keyless trust setup for connected mode, and the switch table in both directions.

## Details worth knowing

- **Half-states refuse before any network call**: `org` without `audience` (or the reverse) fails naming the exact line to add or remove — all problems at once.
- **Image injection is blank-fill**: containers whose image you authored explicitly (sidecars) are never touched; only blank image slots receive the built reference. The `set` input (`<kind>/<name>:<fieldPath>=<value>`, one per line) is the field-exact escape hatch.
- **Runtime secrets never enter git**: manifests carry provider-native secret references (Secret Manager, ECS `valueFrom` ARNs, Key Vault) — references, not values, resolved by your cloud at runtime.
- **Verify locally before the first push**: `PLANTON_BACKEND=none planton service deploy --env prod --preflight-only` prints the same preflight report the runner will see, without deploying anything.
- **Runners**: Linux and macOS, x64 and arm64; the CLI download is checksum-verified before it is ever executed.
