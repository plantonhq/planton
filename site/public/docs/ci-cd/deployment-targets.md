---
title: "Deployment Targets"
description: "Where services run — supported cloud platforms, deployment configuration models, and per-environment targeting."
icon: cloud
order: 40
tags:
  - Deployment
  - Multi-Cloud
  - Service Hub
---

# Deployment Targets

A deployment target defines where a service runs for a given environment. Service Hub supports deploying to multiple cloud platforms, and each environment in your deployment pipeline can target a different provider and resource type.

## Why Two Deployment Models

Service Hub originally supported only Git-based deployment configuration through Kustomize directories. This approach works well for teams comfortable with GitOps workflows, but it creates a split onboarding experience: you configure the service in the web console, then switch to your terminal to create deployment manifests in Git, commit, push, and wait for the pipeline to discover any YAML errors.

The insight that led to inline deployment configuration came from user feedback: the `_kustomize` directory ultimately produces catalog object manifests — the same manifests users already create through Planton's UI for infrastructure deployment. There is no technical reason deployment configuration cannot be captured during service onboarding.

Inline deployment targets close this gap. They enable complete service onboarding through the web console — from Git repository to deployed service — without creating any files in your repository. The choice between Git-based and inline is made per service, not per environment: pick one approach and use it consistently.

## Supported Platforms

Service Hub deploys to any catalog kind marked as service-deployable in the platform. The following targets are available through the web console creation wizard:

| Provider | Resource Type | Description |
|----------|--------------|-------------|
| Kubernetes | Deployment | Standard stateless workload on any Kubernetes cluster (EKS, GKE, AKS, self-managed) |
| Kubernetes | StatefulSet | Stateful workload with persistent storage |
| AWS | ECS Service | Managed container service on Amazon ECS |
| GCP | Cloud Run | Serverless containers on Google Cloud Run |
| Cloudflare | Worker | Edge computing worker (for Cloudflare Worker script artifacts) |

Additional service-deployable types — including DigitalOcean App Platform — are defined in the platform's resource catalog and can be targeted through the Git-based deployment path.

## Git-Based Deployment (Kustomize)

The default model. Deployment configuration lives in a `_kustomize` directory in your Git repository:

```
_kustomize/
  base/
    kustomization.yaml
    deployment.yaml
  overlays/
    dev/
      kustomization.yaml
      patch.yaml
    staging/
      ...
    production/
      ...
```

Each overlay directory represents an environment. During the pipeline's build stage, a kustomize-build task processes each overlay and produces catalog object manifests. The deploy stage then provisions those manifests through Planton's infrastructure layer.

**When to use**: Teams that prefer GitOps workflows, want deployment configuration version-controlled in Git, or need full control over resource manifests with complex per-environment customization.

The web console wizard labels this option **"Git-Based (GitOps)"**.

## Inline Deployment (UI-Based)

Deployment targets are defined directly in the Service configuration. No `_kustomize` directory is needed in the repository.

Each target specifies:

- **Environment**: The deployment environment name (e.g., dev, staging, production)
- **Cloud provider**: Which provider to deploy to (Kubernetes, AWS, GCP, Cloudflare)
- **Resource type**: The specific resource type for this provider (e.g., Deployment, ECS Service, Cloud Run)
- **Resource configuration**: The complete infra component specification as a structured object
- **Manual approval**: Whether the pipeline should pause for approval before deploying to this environment

The resource configuration supports template variables that are substituted during pipeline execution:

- `{{ .Image }}` — The built container image reference
- `{{ .CommitSHA }}` — The Git commit SHA that triggered the pipeline

**When to use**: Getting started quickly, teams that prefer UI-based configuration, or rapid iteration without code changes. The web console wizard labels this option **"UI-Based (Configure Here)"** and marks it as recommended.

<!-- SCREENSHOT: Deployment target configuration wizard
  Page: /orgs/{org}/services (deploy wizard, deployment config step)
  Action: Show the AddTargetModal with environment, provider, and kind selection
  Focus: The multi-step target modal showing environment name, provider cards, and kind dropdown
  Alt: Deployment target wizard showing environment name field, cloud provider selection cards, and resource kind options
-->

## Per-Environment Targeting

A key capability of inline deployment targets is per-environment provider flexibility. Different environments can deploy to different cloud platforms:

```yaml
deployment_targets:
  - env: dev
    provider: kubernetes
    kind: KubernetesDeployment
    catalog_object: { ... }
  - env: staging
    provider: gcp
    kind: GcpCloudRun
    catalog_object: { ... }
  - env: production
    provider: aws
    kind: AwsEcsService
    catalog_object: { ... }
```

This enables incremental migration between cloud providers or using cost-optimized platforms for non-production environments.

## Build-Only Services

Both models support services with no deployment targets. When the deployment target list is empty, the pipeline builds and packages the artifact but skips deployment. This is useful for:

- Services where infrastructure is still being provisioned
- Library packages that only need to be built and published to a registry
- Services in early development where deployment will be configured later

The web console displays a note: "No deployment targets configured. Pipelines will build your artifact but skip deployment."

## Manual Approval Gates

Each deployment target can require manual approval before the pipeline deploys to that environment.

When a manual gate is active:

1. The pipeline completes preceding deployments.
2. The pipeline pauses at the gated environment.
3. A team member approves or rejects via the web console or CLI.
4. On approval, deployment proceeds. On rejection, the pipeline stops.

Manual gates can also be enforced by the organization's promotion policy, independent of per-target settings. If either the target or the policy requires approval, the gate is enforced.

```bash
# Approve or reject a manual gate from the CLI
planton service pipeline resolve-manual-gate <pipeline-id> <deployment-task-name> yes
planton service pipeline resolve-manual-gate <pipeline-id> <deployment-task-name> no
```

## Deployment Ordering

Deployment targets are executed in the order defined by the organization's promotion policy, not by the order in which targets are listed in the Service configuration. The promotion policy defines the canonical deployment sequence (e.g., dev, then staging, then production) and determines which transitions require manual approval.

## Configuring Deployment Targets

### Web Console

During service creation, the deployment configuration step allows adding targets:

1. **Environment name**: Free-text with suggestions (dev, staging, production).
2. **Cloud provider**: Filtered by artifact type — container images show Kubernetes, AWS, GCP; Cloudflare Workers show Cloudflare only.
3. **Resource type**: Provider-specific options (e.g., Deployment or StatefulSet for Kubernetes).
4. **Manual approval**: Optional checkbox per target.

Targets appear as cards that can be edited or removed. After creation, targets are managed in the service details page.

### CLI

Deployment targets are configured as part of the Service spec: edit the service's `service.yaml` and register it with `planton service register -f service.yaml`, which validates the whole manifest before sending it.

For Git-based deployments, manage the `_kustomize` directory structure:

```bash
# Scaffold a _kustomize tree for a new service: one empty overlay per environment, plus the merge schema
planton service kustomize init --envs dev,prod

# Hand an existing service's configuration to the repository as a kustomize tree
planton service kustomize eject <service>
```

## Related Documentation

- [What is a Service?](/docs/ci-cd/what-is-a-service) — Service configuration overview
- [Build Methods](/docs/ci-cd/build-methods) — How artifacts are built
- [Pipelines](/docs/ci-cd/pipelines) — The pipeline execution model including deploy stage
- [Secrets](/docs/secrets) — Secrets management and configuration variables injected into deployments
- [Ingress](/docs/ci-cd/ingress) — Making deployed services accessible
