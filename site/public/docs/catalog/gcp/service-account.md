---
title: "Service Account"
description: "Service Account deployment documentation"
icon: "package"
order: 100
componentName: "gcpserviceaccount"
---

# GCP Service Account

Deploys a Google Cloud service account with optional JSON key generation and IAM role bindings at the project and organization levels. The module creates the account, attaches the specified roles using per-role `IAMMember` bindings, and exports the resulting email address and (when requested) the base64-encoded private key.

## What Gets Created

When you deploy a GcpServiceAccount resource, OpenMCF provisions:

- **Service Account** — a GCP service account in the specified project, with `serviceAccountId` as the account ID and `metadata.name` as the display name
- **Service Account Key** (conditional) — a JSON private key for the service account, created only when `createKey` is set to `true`
- **Project IAM Bindings** — one `projects.IAMMember` resource per entry in `projectIamRoles`, granting each role to the service account in the target project
- **Organization IAM Bindings** — one `organizations.IAMMember` resource per entry in `orgIamRoles`, granting each role to the service account in the specified organization (requires `orgId`)

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** where the service account will be created
- **Organization ID** if you need to assign organization-level IAM roles
- **IAM permissions** to create service accounts and manage IAM bindings in the target project (and organization, if applicable)

## Quick Start

Create a file `service-account.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpServiceAccount
metadata:
  name: my-app-sa
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpServiceAccount.my-app-sa
spec:
  serviceAccountId: my-app-sa
  projectId: my-gcp-project-123
```

Deploy:

```shell
openmcf apply -f service-account.yaml
```

This creates a service account `my-app-sa@my-gcp-project-123.iam.gserviceaccount.com` with no keys and no additional IAM roles.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `serviceAccountId` | `string` | Short unique ID for the service account, used to form the email `<serviceAccountId>@<project>.iam.gserviceaccount.com`. | Required, 6-30 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` or `valueFrom` | Provider default project | The GCP project in which the service account is created. Can be a literal value or a reference to a GcpProject resource. |
| `orgId` | `string` | `""` | GCP organization ID (numeric string). Required when `orgIamRoles` is non-empty. |
| `createKey` | `bool` | `false` | When `true`, a JSON private key is generated for the service account and its base64-encoded value is exported in stack outputs. |
| `projectIamRoles` | `string[]` | `[]` | IAM roles to grant to the service account at the project level (e.g., `roles/logging.logWriter`). One `IAMMember` binding is created per role. |
| `orgIamRoles` | `string[]` | `[]` | IAM roles to grant to the service account at the organization level (e.g., `roles/resourcemanager.organizationViewer`). Requires `orgId` to be set. |

## Examples

### Service Account with Project IAM Roles

A service account with permissions to write logs and read Cloud Storage buckets:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpServiceAccount
metadata:
  name: backend-worker
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpServiceAccount.backend-worker
spec:
  serviceAccountId: backend-worker
  projectId: my-gcp-project-123
  projectIamRoles:
    - roles/logging.logWriter
    - roles/storage.objectViewer
```

### Service Account with Key Generation

A CI/CD service account with a generated JSON key and deployment permissions:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpServiceAccount
metadata:
  name: ci-deployer
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.GcpServiceAccount.ci-deployer
spec:
  serviceAccountId: ci-deployer
  projectId: my-gcp-project-123
  createKey: true
  projectIamRoles:
    - roles/container.developer
    - roles/storage.admin
    - roles/artifactregistry.writer
```

### Service Account with Organization IAM Roles

A service account that needs both project-level and organization-level permissions:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpServiceAccount
metadata:
  name: org-auditor
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpServiceAccount.org-auditor
spec:
  serviceAccountId: org-auditor
  projectId: my-gcp-project-123
  orgId: "123456789012"
  projectIamRoles:
    - roles/logging.logWriter
  orgIamRoles:
    - roles/resourcemanager.organizationViewer
    - roles/iam.securityReviewer
```

### Using a Foreign Key Reference for Project ID

Reference an OpenMCF-managed GcpProject instead of hardcoding the project ID:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpServiceAccount
metadata:
  name: app-runtime
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpServiceAccount.app-runtime
spec:
  serviceAccountId: app-runtime
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  createKey: true
  projectIamRoles:
    - roles/cloudrun.invoker
    - roles/secretmanager.secretAccessor
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `email` | `string` | The full email address of the created service account (`<serviceAccountId>@<project>.iam.gserviceaccount.com`) |
| `keyBase64` | `string` | Base64-encoded JSON private key. Only populated when `createKey` is `true`. |

## Related Components

- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project where the service account is created
- [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — GKE clusters that may use this service account for workload identity
- [GcpDnsZone](/docs/catalog/gcp/dns-zone) — DNS zones that can reference service accounts in `iamServiceAccounts`
- [GcpCloudRun](/docs/catalog/gcp/cloud-run) — Cloud Run services that may run under this service account
