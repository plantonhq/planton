---
title: "Multi-Environment Deployments"
description: "Deploy the same PostgreSQL component to dev, staging, and production with environment-specific configuration using Kustomize overlays"
icon: "tutorial"
order: 40
---

# Multi-Environment Deployments

In this tutorial, you will deploy the same PostgreSQL database component to three environments — dev, staging, and production — with different resource sizing, replica counts, and configuration for each. Instead of maintaining three separate manifests, you will use Kustomize overlays to define a shared base and environment-specific patches.

By the end, you will know how to structure a Kustomize directory, write overlay patches, and deploy to any environment with a single flag change.

## What You Will Build

Three PostgreSQL deployments from one base manifest:

| Environment | Replicas | CPU Limit | Memory Limit | Disk | Ingress |
|-------------|----------|-----------|--------------|------|---------|
| dev | 1 | 500m | 512Mi | 1Gi | disabled |
| staging | 2 | 1000m | 1Gi | 5Gi | disabled |
| prod | 3 | 2000m | 4Gi | 20Gi | enabled |

## Prerequisites

Before starting, ensure you have:

- **Planton CLI** installed (`planton version`). See [Getting Started](../getting-started) for installation.
- **A Kubernetes cluster** running and accessible via `kubectl`.
- **Pulumi CLI** installed with a backend configured.
- Familiarity with the [first Kubernetes resource tutorial](./first-kubernetes-resource) — this tutorial builds on that manifest.

Kustomize itself does not need to be installed separately. Planton embeds Kustomize as a Go library and runs it internally.

## Step 1: Create the Directory Structure

Set up the standard Kustomize layout:

```bash
mkdir -p postgres-kustomize/base
mkdir -p postgres-kustomize/overlays/dev
mkdir -p postgres-kustomize/overlays/staging
mkdir -p postgres-kustomize/overlays/prod
```

The final structure will look like this:

```text
postgres-kustomize/
|-- base/
|   |-- kustomization.yaml
|   \-- postgres.yaml
\-- overlays/
    |-- dev/
    |   |-- kustomization.yaml
    |   \-- patch.yaml
    |-- staging/
    |   |-- kustomization.yaml
    |   \-- patch.yaml
    \-- prod/
        |-- kustomization.yaml
        \-- patch.yaml
```

## Step 2: Write the Base Manifest

The base manifest defines the shared configuration that all environments inherit.

**`postgres-kustomize/base/postgres.yaml`**:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: app-database
  labels:
    planton.dev/provisioner: pulumi
spec:
  namespace:
    value: app-database
  createNamespace: true
  container:
    replicas: 1
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 500m
        memory: 512Mi
    diskSize: 1Gi
  databases:
    - name: appdb
      ownerRole: appuser
  users:
    - name: appuser
      flags:
        - login
        - createdb
  ingress:
    enabled: false
```

**`postgres-kustomize/base/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - postgres.yaml
```

The base uses conservative defaults — small resources, single replica, no ingress. Overlays scale up from here.

## Step 3: Write the Dev Overlay

Dev uses the base as-is with only a label change to identify the environment.

**`postgres-kustomize/overlays/dev/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - path: patch.yaml
```

**`postgres-kustomize/overlays/dev/patch.yaml`**:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: app-database
  labels:
    environment: dev
```

Dev inherits all base values. The patch only adds an environment label.

## Step 4: Write the Staging Overlay

Staging increases replicas, resources, and disk to test with production-like sizing.

**`postgres-kustomize/overlays/staging/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - path: patch.yaml
```

**`postgres-kustomize/overlays/staging/patch.yaml`**:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: app-database
  labels:
    environment: staging
spec:
  container:
    replicas: 2
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1000m
        memory: 1Gi
    diskSize: 5Gi
```

Only the fields that differ from the base are specified. Kustomize performs a strategic merge — everything else (databases, users, namespace, ingress) comes from the base.

## Step 5: Write the Production Overlay

Production maximizes resources, adds replicas for high availability, and enables ingress for external access.

**`postgres-kustomize/overlays/prod/kustomization.yaml`**:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - path: patch.yaml
```

**`postgres-kustomize/overlays/prod/patch.yaml`**:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: app-database
  labels:
    environment: prod
spec:
  container:
    replicas: 3
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        cpu: 2000m
        memory: 4Gi
    diskSize: 20Gi
  ingress:
    enabled: true
    hostname: postgres.example.com
```

The production patch enables ingress with a hostname. The ingress validation rule in the component's Protocol Buffer schema requires `hostname` when `enabled` is `true` — omitting it would fail validation.

## Step 6: Deploy to Dev

Preview the dev deployment:

```bash
planton plan --kustomize-dir postgres-kustomize --overlay dev
```

Planton builds the Kustomize output by merging `base/postgres.yaml` with `overlays/dev/patch.yaml`, validates the result against the `KubernetesPostgres` schema, and generates the execution plan.

Deploy:

```bash
planton apply --kustomize-dir postgres-kustomize --overlay dev
```

Verify:

```bash
kubectl get pods -n app-database
```

You should see a single PostgreSQL pod running with the dev configuration.

## Step 7: Deploy to Staging

```bash
planton plan --kustomize-dir postgres-kustomize --overlay staging
```

Review the plan — you should see 2 replicas and larger resource allocations compared to dev.

```bash
planton apply --kustomize-dir postgres-kustomize --overlay staging
```

If deploying to the same cluster, note that both environments share the same namespace name (`app-database`). In a real setup, you would either use separate clusters per environment or customize the namespace in each overlay patch (e.g., `app-database-staging`).

## Step 8: Compare Environments

The power of the overlay approach is visible when you compare what each environment deploys. The base manifest stays identical — only the patches differ:

| What changes | Dev | Staging | Prod |
|-------------|-----|---------|------|
| Labels | `environment: dev` | `environment: staging` | `environment: prod` |
| Replicas | 1 (base) | 2 | 3 |
| CPU limit | 500m (base) | 1000m | 2000m |
| Memory limit | 512Mi (base) | 1Gi | 4Gi |
| Disk | 1Gi (base) | 5Gi | 20Gi |
| Ingress | disabled (base) | disabled (base) | enabled |
| Databases | appdb (base) | appdb (base) | appdb (base) |
| Users | appuser (base) | appuser (base) | appuser (base) |

Adding a new database or user in the base manifest automatically propagates to all environments without touching any overlay.

## Step 9: Clean Up

Destroy each environment:

```bash
planton destroy --kustomize-dir postgres-kustomize --overlay dev
planton destroy --kustomize-dir postgres-kustomize --overlay staging
```

Each destroy command uses the same Kustomize flags to resolve the correct manifest for teardown.

## What You Learned

- How to structure a Kustomize directory with a shared base and per-environment overlays
- How strategic merge patches override only the fields that differ per environment
- How the `--kustomize-dir` and `--overlay` flags work together to build the final manifest
- How validation rules (like requiring `hostname` when ingress is enabled) apply to the merged result, not individual patches
- How shared configuration (databases, users) defined in the base propagates to all environments automatically

## What's Next

- [Kustomize Integration](../guides/kustomize) — advanced patterns including JSON 6902 patches, components, multiple bases, and ConfigMap generators
- [CI/CD Integration](../guides/cicd-integration) — automating multi-environment deployments with branch-based overlay selection in GitHub Actions and GitLab CI
- [Deploy Across Providers](./multi-provider) — deploy the same resource type on AWS and GCP
- [State Backends](../guides/state-backends) — configure per-environment state storage
