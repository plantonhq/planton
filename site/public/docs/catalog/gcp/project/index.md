---
title: "Project"
description: "Project deployment documentation"
icon: "package"
order: 100
componentName: "gcpproject"
---

# GCP Project

Creates and configures a Google Cloud project within your resource hierarchy. The component handles project creation under an organization or folder, billing account linkage, standard label propagation, optional removal of the default VPC network, API enablement, and an optional owner IAM binding.

## What Gets Created

When you deploy a GcpProject resource, Planton provisions:

- **GCP Project** — a `google_project` resource placed under the specified organization or folder, with billing account attached and GCP labels applied
- **Random Suffix** (conditional) — a 3-character lowercase alphabetic suffix appended to `projectId` when `addSuffix` is `true`, ensuring uniqueness across deployments
- **Default Network Removal** (conditional) — sets `auto_create_network` to `false` when `disableDefaultNetwork` is `true` (the default), preventing GCP from provisioning the insecure default VPC
- **Enabled APIs** — a `google_project_service` resource for each entry in `enabledApis`, activating the specified Cloud APIs on the new project
- **Owner IAM Binding** (conditional) — a `google_project_iam_member` resource granting `roles/owner` to the principal specified in `ownerMember`, created only when that field is set
- **Deletion Protection** (conditional) — sets the project deletion policy to `PREVENT` when `deleteProtection` is `true`, blocking accidental project deletion at the GCP level

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP organization or folder** — referenced via `parentType` and `parentId`
- **IAM permissions** to create projects under the target organization or folder (`roles/resourcemanager.projectCreator`)
- **A billing account** — referenced via `billingAccountId` if the project will consume billable services

## Quick Start

Create a file `project.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpProject
metadata:
  name: my-project
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpProject.my-project
spec:
  projectId: my-dev-project-01
  parentType: organization
  parentId: "123456789012"
  billingAccountId: 0123AB-4567CD-89EFGH
```

Deploy:

```shell
planton apply -f project.yaml
```

This creates a GCP project named `my-project` with project ID `my-dev-project-01` under the specified organization, with the default VPC network removed and billing linked.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `string` | Globally unique GCP project ID. | 6-30 chars; lowercase letters, digits, hyphens; must start with a letter; cannot end with a hyphen |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `addSuffix` | `bool` | `false` | When `true`, appends a random 3-character lowercase suffix to `projectId` to ensure uniqueness. Useful for temporary or test projects. |
| `parentType` | `enum` | — | Type of resource hierarchy parent: `organization` or `folder`. |
| `parentId` | `string` | — | Numeric ID of the parent organization or folder. |
| `billingAccountId` | `string` | — | Billing account in the form `0123AB-4567CD-89EFGH`. Required for any project using billable services. |
| `labels` | `map<string, string>` | — | Key/value metadata labels for cost allocation and governance. Keys must be lowercase letters, digits, or underscores (max 63 chars). |
| `disableDefaultNetwork` | `bool` | `true` | When `true`, prevents GCP from auto-creating the default VPC network. This is a common security hardening step. |
| `enabledApis` | `string[]` | — | List of Cloud APIs to enable (e.g., `compute.googleapis.com`). Each entry must end with `.googleapis.com`. |
| `ownerMember` | `string` | — | IAM member (user, group, or service account email) to grant `roles/owner` at project creation. |
| `deleteProtection` | `bool` | `false` | When `true`, sets the GCP-native deletion policy to `PREVENT`, blocking project deletion until this flag is disabled. |

## Examples

### Minimal Project Under an Organization

A basic project with billing and default security hardening (no default VPC):

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpProject
metadata:
  name: sandbox
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpProject.sandbox
spec:
  projectId: acme-sandbox-dev
  parentType: organization
  parentId: "112233445566"
  billingAccountId: 0123AB-4567CD-89EFGH
```

### Project Under a Folder with Enabled APIs

A staging project placed under a folder, with Compute Engine, Cloud Run, and Artifact Registry APIs pre-enabled:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpProject
metadata:
  name: staging-project
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.GcpProject.staging-project
spec:
  projectId: acme-staging-web
  parentType: folder
  parentId: "998877665544"
  billingAccountId: 0123AB-4567CD-89EFGH
  labels:
    team: platform
    cost_center: eng_staging
  enabledApis:
    - compute.googleapis.com
    - run.googleapis.com
    - artifactregistry.googleapis.com
```

### Production Project with Owner, Deletion Protection, and Random Suffix

A production project with all optional features: a designated owner, deletion protection to prevent accidental removal, a random suffix for uniqueness, and a full set of enabled APIs:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpProject
metadata:
  name: prod-data
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpProject.prod-data
spec:
  projectId: acme-prod-data
  addSuffix: true
  parentType: organization
  parentId: "112233445566"
  billingAccountId: 0123AB-4567CD-89EFGH
  labels:
    team: data-engineering
    cost_center: eng_prod
    compliance: soc2
  disableDefaultNetwork: true
  deleteProtection: true
  ownerMember: group:devops-admins@acme.com
  enabledApis:
    - compute.googleapis.com
    - container.googleapis.com
    - sqladmin.googleapis.com
    - secretmanager.googleapis.com
    - servicenetworking.googleapis.com
    - cloudresourcemanager.googleapis.com
```

Because `addSuffix` is `true`, the final project ID will be something like `acme-prod-data-xkf`, ensuring no collision if this manifest is applied in multiple stacks.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | Display name of the project (mirrors `metadata.name`). |
| `project_id` | `string` | Immutable project ID. When `addSuffix` is `true`, this includes the generated suffix. |
| `project_number` | `string` | Numeric project number assigned by Google. Used by some APIs and IAM policies that reference projects by number. |

## Related Components

- [GcpVpc](/docs/catalog/gcp/vpc) — creates a VPC network inside this project
- [GcpSubnetwork](/docs/catalog/gcp/subnetwork) — creates subnets within a VPC in this project
- [GcpServiceAccount](/docs/catalog/gcp/service-account) — provisions service accounts scoped to this project
- [GcpSecretsManager](/docs/catalog/gcp/secrets-manager) — manages secrets in this project using Secret Manager
- [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — deploys a GKE cluster into this project
- [GcpCloudSql](/docs/catalog/gcp/cloud-sql) — deploys Cloud SQL instances in this project
- [GcpDnsZone](/docs/catalog/gcp/dns-zone) — creates Cloud DNS managed zones in this project
