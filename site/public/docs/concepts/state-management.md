---
title: "State Management"
description: "How Planton manages deployment state across Pulumi and OpenTofu/Terraform backends -- from Pulumi Cloud and S3 to GCS, Azure Blob, and local filesystem"
icon: "database"
order: 50
---

# State Management

Every IaC deployment tracks state -- the mapping between what your manifest declares and what actually exists in the cloud. State enables Planton to know what has changed between deployments, what needs to be created, updated, or destroyed, and what the current outputs of a deployment are.

How state is stored depends on which IaC engine you use. Pulumi and OpenTofu/Terraform have different backend systems, and Planton configures them through manifest labels.

## Pulumi State Backends

Pulumi organizes state by **stack** -- a named instance of a deployment identified by a fully qualified domain name (FQDN) in the format `{organization}/{project}/{stack-name}`.

### Configuring Pulumi State

Pulumi state configuration is provided through manifest labels:

```yaml
metadata:
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: platform
    pulumi.planton.dev/stack.name: production.KubernetesPostgres.session-store
```

Or as a single FQDN using the `--stack` flag:

```bash
planton pulumi up -f postgres.yaml --stack acme/platform/production
```

### Supported Pulumi Backends

| Backend | Description |
|---------|-------------|
| **Pulumi Cloud** | Managed state storage with history, audit trail, and team collaboration. Default backend. |
| **S3** | Self-managed state in an AWS S3 bucket. |
| **GCS** | Self-managed state in a Google Cloud Storage bucket. |
| **Azure Blob** | Self-managed state in Azure Blob Storage. |
| **Local** | State stored on the local filesystem. Suitable for development only. |

The Pulumi backend is configured through Pulumi's standard mechanisms (environment variables, `pulumi login`, or configuration files). Planton's role is to set the stack FQDN from the manifest labels, not to configure the backend connection itself.

## OpenTofu/Terraform State Backends

OpenTofu and Terraform use a `backend` block in the Terraform configuration to determine where state is stored. Planton generates this configuration from manifest labels.

### Configuring Tofu/Terraform State

State backend configuration is provided through manifest labels with the `tofu.planton.dev/backend.*` prefix:

```yaml
metadata:
  labels:
    planton.dev/provisioner: tofu
    tofu.planton.dev/backend.type: s3
    tofu.planton.dev/backend.bucket: my-tfstate-bucket
    tofu.planton.dev/backend.key: prod/postgres/terraform.tfstate
    tofu.planton.dev/backend.region: us-east-1
```

The CLI reads these labels, writes a `backend.tf` file in the workspace, and passes the configuration to `tofu init`.

Legacy `terraform.planton.dev/backend.*` labels are also supported for backward compatibility.

### Supported Backends and Required Fields

**S3**

| Label | Required | Description | Example |
|-------|----------|-------------|---------|
| `backend.type` | Yes | Must be `s3` | `s3` |
| `backend.bucket` | Yes | S3 bucket name | `my-terraform-state-bucket` |
| `backend.key` | Yes | State file path within bucket | `env/prod/terraform.tfstate` |
| `backend.region` | Yes | AWS region (or `auto` for S3-compatible) | `us-west-2` |
| `backend.endpoint` | Only if `region=auto` | Custom S3-compatible endpoint | `https://acct.r2.cloudflarestorage.com` |

**GCS**

| Label | Required | Description | Example |
|-------|----------|-------------|---------|
| `backend.type` | Yes | Must be `gcs` | `gcs` |
| `backend.bucket` | Yes | GCS bucket name | `my-terraform-state` |
| `backend.key` | Yes | Prefix path within bucket | `terraform/state` |

**Azure Storage (azurerm)**

| Label | Required | Description | Example |
|-------|----------|-------------|---------|
| `backend.type` | Yes | Must be `azurerm` | `azurerm` |
| `backend.bucket` | Yes | Azure Storage container name | `tfstate` |
| `backend.key` | Yes | State file blob name | `prod.terraform.tfstate` |

**Local**

| Label | Required | Description |
|-------|----------|-------------|
| `backend.type` | Yes | Must be `local` |

No other fields are required for local backends. State is stored in the workspace directory.

### S3-Compatible Backends (Cloudflare R2, MinIO)

Planton supports S3-compatible backends for teams using Cloudflare R2, MinIO, or other S3-compatible object stores. To use an S3-compatible backend, set the region to `auto` and provide the custom endpoint:

```yaml
metadata:
  labels:
    tofu.planton.dev/backend.type: s3
    tofu.planton.dev/backend.bucket: my-r2-state
    tofu.planton.dev/backend.key: prod/terraform.tfstate
    tofu.planton.dev/backend.region: auto
    tofu.planton.dev/backend.endpoint: https://your-account-id.r2.cloudflarestorage.com
```

When `region=auto` is detected, the CLI automatically configures the S3 backend with the compatibility flags that S3-compatible backends require.

## State Lifecycle

Both engines follow the same conceptual lifecycle:

| Phase | Pulumi Command | Tofu/Terraform Command | What Happens |
|-------|---------------|----------------------|--------------|
| **Initialize** | `planton pulumi init` | `planton tofu init` | Set up backend, download providers |
| **Preview** | `planton pulumi preview` | `planton tofu plan` | Compare manifest against state, show planned changes |
| **Apply** | `planton pulumi up` | `planton tofu apply` | Execute changes, update state |
| **Refresh** | `planton pulumi refresh` | `planton tofu refresh` | Sync state with actual cloud resources |
| **Destroy** | `planton pulumi destroy` | `planton tofu destroy` | Delete all resources, clean up state |

The preview/plan step is where state management provides its key value: by comparing your manifest against the stored state, the engine can tell you exactly what will change before you commit to the deployment.

## Multi-Environment State Isolation

Each environment should have its own isolated state. In Pulumi, this is achieved through different stack names:

```yaml
# Production
pulumi.planton.dev/stack.name: production.KubernetesPostgres.session-store

# Staging
pulumi.planton.dev/stack.name: staging.KubernetesPostgres.session-store
```

In OpenTofu/Terraform, this is achieved through different state file keys:

```yaml
# Production
tofu.planton.dev/backend.key: production/postgres/terraform.tfstate

# Staging
tofu.planton.dev/backend.key: staging/postgres/terraform.tfstate
```

Both approaches ensure that a deployment to staging never reads or modifies the production state, and vice versa.

## What's Next

- **[Dual IaC Engines](dual-iac-engines)** -- How the Pulumi and OpenTofu/Terraform engines work
- **[Manifests](manifests)** -- How manifest labels configure state backends
- **[Module System](module-system)** -- How IaC modules are resolved before state operations run
