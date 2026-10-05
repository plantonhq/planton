# GCP Project IAM Member - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying additive project-level IAM grants using Planton's `GcpProjectIamMember` API. The module is written in Go and uses the Pulumi GCP provider to create `projects.IAMMember` resources (backed by `google_project_iam_member`).

The grant is ADDITIVE: it merges one (role, member[, condition]) pair into the project's IAM policy without touching any other member's bindings, and destroy subtracts only this exact pair.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** — the grant is policy metadata; no billed infrastructure is created
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── README.md         # This file
└── module/
    ├── main.go        # Module coordinator
    ├── iam_member.go  # IAM member grant creation
    ├── locals.go      # Resolved resource holder
    └── outputs.go     # Output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `iac-input.yaml` with the grant specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpProjectIamMember
  metadata:
    name: app-logs-writer-grant
  spec:
    project_id:
      value: my-gcp-project-123
    role:
      value: roles/logging.logWriter
    member:
      value: serviceAccount:my-app@my-gcp-project-123.iam.gserviceaccount.com
```

### 3. Build, Preview, and Deploy

```bash
make build
pulumi preview
pulumi up
```

### 4. View Outputs

```bash
pulumi stack output role
pulumi stack output member
pulumi stack output etag
```

## Inputs

The module consumes `GcpProjectIamMemberIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpProjectIamMember` spec (role, member, optional project/condition) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

Spec fields: `role` and `member` (both required; usually references to GcpIamCustomRole / GcpServiceAccount outputs), optional `project_id` (empty falls back to the provider's default project, resolved concretely via the provider client config), optional `condition` (title + expression required).

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `project_id` | string | The project whose policy received the grant |
| `role` | string | The granted role |
| `member` | string | The granted member |
| `etag` | string | The project IAM policy etag after the grant |

## Behavior Notes

- **Everything is ForceNew**: IAM grants have no update; any change replaces the grant atomically, mirroring the underlying API.
- **Member format is validated at deploy time** (not in the proto) because the value usually arrives through a reference resolved only at deploy time. Deleted principals are rejected.
- **Conditions are part of the grant's identity**: the same role with and without a condition are two independent policy entries.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
