---
title: "Resource Browser"
description: "Browse and list infra components across AWS, GCP, and Azure from a single CLI without distributing cloud console credentials"
icon: cloud
order: 30
tags:
  - Cloud Ops
  - AWS
  - GCP
  - Azure
  - Multi-Cloud
---

# Resource Browser

Beyond Kubernetes, Cloud Ops provides resource listing operations across AWS, GCP, and Azure. List EC2 instances, S3 buckets, Compute Engine VMs, Cloud Storage buckets, Azure VMs, and Blob Storage containers — all from the Planton CLI, all routed through the secure tunnel without distributing cloud console credentials.

## Why It Matters

Infrastructure administrators typically need to check resource state across multiple cloud accounts and providers — a quick look at which EC2 instances are running, how many storage buckets exist, or whether Azure VMs are healthy. Without Cloud Ops, this means switching between three cloud consoles, each with their own credentials and IAM policies.

Cloud Ops collapses this into a single CLI interface. The credentials stay in customer infrastructure (on the Planton Runner), and the administrator's access is controlled through Planton's connection authorization model rather than per-provider IAM.

<!-- SCREENSHOT: Resource browser CLI output
  Page: Terminal / CLI
  Action: Show the output of a planton aws ec2 describe-instances command in table format
  Focus: The tabular output showing instance IDs, types, states, and regions
  Alt: CLI output of planton aws ec2 describe-instances showing EC2 instances in table format with status indicators
-->

## AWS Operations

### List EC2 Instances

List EC2 instances with optional filters:

```bash
planton aws ec2 describe-instances --connection aws-prod --region us-east-1
```

The `describe-instances` command (also aliased as `ls`) accepts the following flags:

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--region` | Yes | — | AWS region to query |
| `--connection` | No | — | AWS connection slug (uses default if empty) |
| `--env` | No | — | Environment slug for default connection resolution |
| `--instance-ids` | No | — | Specific instance IDs to describe |
| `--filters` | No | — | Filter expressions in `Name=value` format |
| `--output` | No | table | Output format: `table`, `json`, `yaml` |

Filter examples:

```bash
# List only running instances
planton aws ec2 ls --connection aws-prod --region us-east-1 \
  --filters "instance-state-name=running"

# Filter by instance type
planton aws ec2 ls --connection aws-prod --region us-west-2 \
  --filters "instance-type=t3.medium"
```

### List S3 Buckets

List all S3 buckets accessible through a connection:

```bash
planton aws s3 ls --connection aws-prod
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--connection` | No | — | AWS connection slug (uses default if empty) |
| `--env` | No | — | Environment slug for default connection resolution |
| `--region` | No | — | AWS region for SDK client initialization |
| `--output` | No | table | Output format: `table`, `json`, `yaml` |

## GCP Operations

The GCP commands use `planton gcp` (or the alias `planton gcloud`).

### List Compute Engine Instances

List VM instances in a project:

```bash
planton gcp compute instances list --connection gcp-prod --project my-gcp-project
```

The `list` command (also aliased as `ls`) accepts the following flags:

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--project` | Yes | — | GCP project ID |
| `--connection` | No | — | GCP connection slug (uses default if empty) |
| `--env` | No | — | Environment slug for default connection resolution |
| `--zone` | No | - (all zones) | GCP zone to query |
| `--filter` | No | — | GCP filter expression |
| `--output` | No | table | Output format: `table`, `json`, `yaml` |

Filter examples:

```bash
# List only running instances
planton gcp compute instances ls --connection gcp-prod --project my-project \
  --filter "status=RUNNING"

# Filter by label
planton gcp compute instances ls --connection gcp-prod --project my-project \
  --filter "labels.env=production"
```

### List Cloud Storage Buckets

List storage buckets in a project:

```bash
planton gcp storage buckets list --connection gcp-prod --project my-gcp-project
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--project` | Yes | — | GCP project ID |
| `--connection` | No | — | GCP connection slug (uses default if empty) |
| `--env` | No | — | Environment slug for default connection resolution |
| `--prefix` | No | — | Filter bucket names by prefix |
| `--output` | No | table | Output format: `table`, `json`, `yaml` |

## Azure Operations

The Azure commands use `planton azure` (or the alias `planton az`).

### List Virtual Machines

List VMs in an Azure subscription:

```bash
planton azure vm list --connection azure-prod --subscription my-subscription-id
```

The `list` command (also aliased as `ls`) accepts the following flags:

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--subscription` | Yes | — | Azure subscription ID |
| `--connection` | No | — | Azure connection slug (uses default if empty) |
| `--env` | No | — | Environment slug for default connection resolution |
| `-g, --resource-group` | No | — | Filter VMs to a specific resource group |
| `--output` | No | table | Output format: `table`, `json`, `yaml` |

### List Blob Storage Containers

List blob containers in an Azure Storage account:

```bash
planton azure storage container list --connection azure-prod \
  --subscription my-subscription-id \
  -g my-resource-group \
  --storage-account mystorageaccount
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--subscription` | Yes | — | Azure subscription ID |
| `-g, --resource-group` | Yes | — | Resource group containing the storage account |
| `--storage-account` | Yes | — | Azure Storage Account name |
| `--connection` | No | — | Azure connection slug (uses default if empty) |
| `--env` | No | — | Environment slug for default connection resolution |
| `--prefix` | No | — | Filter container names by prefix |
| `--output` | No | table | Output format: `table`, `json`, `yaml` |

## Connection Resolution

All resource browser commands follow the same connection resolution pattern. When you run a command, Cloud Ops determines which credential to use:

1. **Explicit connection** — If `--connection` is provided, use that credential directly
2. **Environment default** — If `--connection` is empty but `--env` is provided, look up the default connection for that environment and provider
3. **Organization default** — If neither is provided, look up the organization-level default connection for the provider

This means you can skip the `--connection` flag entirely if you have [default connections](/docs/connections/default-connections) configured:

```bash
# Uses the org-level default AWS connection
planton aws s3 ls

# Uses the environment-level default AWS connection for prod
planton aws ec2 ls --env prod --region us-east-1
```

## Current Scope

AWS, GCP, and Azure operations currently support resource listing. Kubernetes has the full operation set — pod management, log streaming, exec, resource editing, and deletion.

The Cloud Ops architecture is provider-extensible. The same tunnel routing, connection resolution, and authorization model applies across all providers. Richer operation sets for non-Kubernetes providers will be added based on demand.

## Related Documentation

- [Operations Overview](/docs/operations) — What Cloud Ops is, dual access modes, how the tunnel works
- [Kubernetes Operations](/docs/operations/kubernetes-operations) — Full Kubernetes operations reference
- [Connections > Cloud Providers](/docs/connections/cloud-providers) — How AWS, GCP, and Azure credentials are managed
- [Connections > Default Connections](/docs/connections/default-connections) — How default connection resolution works
