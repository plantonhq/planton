# GCP IAM Custom Role - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP IAM custom roles using Planton's `GcpIamCustomRole` API. The module is written in Go and uses the Pulumi GCP provider to create `projects.IAMCustomRole` resources (backed by `google_project_iam_custom_role`).

A custom role is a named, least-privilege permission bundle: define it once, grant it with GcpProjectIamMember wherever it applies. Permission edits update the role in place and propagate immediately to every existing grant.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** — custom roles are project metadata; no billed infrastructure is created
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: the deploying principal's permissions are listed in [`../permissions.yaml`](../permissions.yaml)

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── README.md         # This file
└── module/
    ├── main.go         # Module coordinator
    ├── custom_role.go  # Custom role resource creation
    ├── locals.go       # Resolved resource holder
    └── outputs.go      # Output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `iac-input.yaml` with the custom role specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpIamCustomRole
  metadata:
    name: log-bucket-writer
  spec:
    project_id:
      value: my-gcp-project-123
    role_id: logBucketWriter
    title: Log Bucket Writer
    permissions:
      - storage.objects.create
      - storage.objects.get
```

### 3. Build, Preview, and Deploy

```bash
make build
pulumi preview
pulumi up
```

### 4. View Outputs

```bash
pulumi stack output name
pulumi stack output role_id
```

## Inputs

The module consumes `GcpIamCustomRoleIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpIamCustomRole` spec (role_id, title, permissions, etc.) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

Spec fields: `role_id`, `project_id` (falls back to the provider default project when empty), `title`, optional `description`, `permissions` (min 1), optional `stage` (default GA).

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `name` | string | Fully-qualified role name (`projects/<project>/roles/<role_id>`) — the grantable handle |
| `role_id` | string | The bare role ID within the project |
| `deleted` | bool | Whether the role is currently soft-deleted |

## Behavior Notes

- **Immutability**: `role_id` and `project_id` are ForceNew; everything else updates in place.
- **Soft delete**: destroying the role soft-deletes it in GCP for up to 14 days. Re-creating a role with a soft-deleted ID undeletes and patches it — the provider handles this natively, so destroy/recreate cycles converge.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
