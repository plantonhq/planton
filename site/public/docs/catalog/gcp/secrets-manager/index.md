---
title: "Secrets Manager"
description: "Secrets Manager deployment documentation"
icon: "package"
order: 100
componentName: "gcpsecretsmanager"
---

# GCP Secrets Manager

Deploys secrets in Google Cloud Secret Manager with automatic replication and placeholder secret versions. Each secret name in the spec produces a dedicated Secret resource and an initial SecretVersion, with secret IDs optionally prefixed by the environment label for multi-environment isolation.

## What Gets Created

When you deploy a GcpSecretsManager resource, Planton provisions:

- **Secret** — one `secretmanager.Secret` per entry in `secretNames`, created in the specified GCP project with automatic replication enabled and Planton-managed labels (resource kind, name, organization, environment)
- **Secret Version** (placeholder) — one `secretmanager.SecretVersion` per secret, initialized with a placeholder value and configured with `ignoreChanges` on `secretData` so that subsequent manual or programmatic updates to the secret value are never overwritten by Planton
- **Environment-Prefixed IDs** — when `metadata.env` is set, each secret ID is formatted as `{env}-{secretName}`; otherwise the secret ID matches the secret name exactly

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** where the secrets will be created
- **Secret Manager API** enabled on the target GCP project (`secretmanager.googleapis.com`)
- **IAM permissions** to create secrets and secret versions in the target project (e.g., `roles/secretmanager.admin`)

## Quick Start

Create a file `secrets.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpSecretsManager
metadata:
  name: my-app-secrets
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpSecretsManager.my-app-secrets
spec:
  projectId: my-gcp-project-123
  secretNames:
    - database-password
```

Deploy:

```shell
planton apply -f secrets.yaml
```

This creates a single secret named `database-password` in Google Cloud Secret Manager with a placeholder version. Update the secret value through the GCP Console, `gcloud`, or application code after deployment.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `string` or `valueFrom` | The GCP project ID where the secrets will be created. Can be a literal value or a reference to a GcpProject resource. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `secretNames` | `string[]` | `[]` | A list of secret names to create in Google Cloud Secret Manager. Each name produces one Secret resource and one placeholder SecretVersion. Empty strings in the list are skipped. |

## Examples

### Single Secret for a Database Password

A minimal deployment that creates one secret for storing a database credential:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpSecretsManager
metadata:
  name: db-credentials
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpSecretsManager.db-credentials
spec:
  projectId: my-gcp-project-123
  secretNames:
    - db-password
```

### Multiple Secrets for a Microservice

Create several secrets that a backend service needs at runtime:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpSecretsManager
metadata:
  name: payment-service-secrets
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.GcpSecretsManager.payment-service-secrets
  env: staging
spec:
  projectId: my-gcp-project-123
  secretNames:
    - stripe-api-key
    - stripe-webhook-secret
    - database-url
    - jwt-signing-key
```

Because `metadata.env` is `staging`, the resulting secret IDs in GCP will be `staging-stripe-api-key`, `staging-stripe-webhook-secret`, `staging-database-url`, and `staging-jwt-signing-key`.

### Production Secrets with Foreign Key Reference

Reference an Planton-managed GcpProject instead of hardcoding the project ID, suitable for a production environment with full application secrets:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpSecretsManager
metadata:
  name: platform-secrets
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpSecretsManager.platform-secrets
  env: prod
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  secretNames:
    - database-password
    - redis-auth-token
    - smtp-credentials
    - oauth-client-secret
    - encryption-master-key
    - third-party-api-key
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `secretIdMap` | `map<string, string>` | A map where each key is the original secret name from `secretNames` and each value is the resolved secret ID in GCP Secret Manager. When `metadata.env` is set, the secret ID is `{env}-{secretName}`; otherwise it equals the secret name. |

## Related Components

- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project where secrets are created
- [GcpServiceAccount](/docs/catalog/gcp/service-account) — creates service accounts that can be granted `roles/secretmanager.secretAccessor` to read secret values at runtime
- [GcpCloudRun](/docs/catalog/gcp/cloud-run) — Cloud Run services that consume secrets as environment variables or mounted volumes
- [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — GKE clusters whose workloads can access secrets via workload identity
