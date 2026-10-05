# GCP Workload Identity Pool - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP Workload Identity Pools using Planton's `GcpWorkloadIdentityPool` API. The module is written in Go and uses the Pulumi GCP provider to create `iam.WorkloadIdentityPool` resources (backed by `google_iam_workload_identity_pool`).

A pool is the trust boundary for keyless authentication: external identities federate through it instead of holding service-account keys. The pool holds no issuer configuration — attach one GcpWorkloadIdentityPoolProvider per issuer.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** — pools are IAM metadata; no billed infrastructure is created
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
    ├── main.go                     # Module coordinator
    ├── workload_identity_pool.go   # Pool resource creation
    ├── locals.go                   # Resolved resource holder
    └── outputs.go                  # Output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `iac-input.yaml` with the pool specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpWorkloadIdentityPool
  metadata:
    name: github-actions-pool
  spec:
    project_id:
      value: my-gcp-project-123
    workload_identity_pool_id: github-actions
    display_name: GitHub Actions
    description: Keyless federation for the engineering org's CI pipelines
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
pulumi stack output workload_identity_pool_id
```

## Inputs

The module consumes `GcpWorkloadIdentityPoolIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpWorkloadIdentityPool` spec (pool ID, display name, mode, etc.) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

Spec fields: `workload_identity_pool_id` (required, immutable), `project_id` (falls back to the provider default project when empty), optional `display_name`, `description`, `disabled`, `mode` (default `FEDERATION_ONLY`), and the trust-domain surface (`inline_certificate_issuance_config`, `inline_trust_config`).

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `name` | string | Full resource name — the handle IAM principals and providers are built from |
| `workload_identity_pool_id` | string | The bare pool ID (providers reference this) |
| `state` | string | `ACTIVE`, or `DELETED` while soft-deleted |

## Behavior Notes

- **Immutability**: `workload_identity_pool_id`, `project_id`, and `mode` are ForceNew (the API rejects mode updates outright); everything else updates in place.
- **Soft delete without undelete**: destroying the pool soft-deletes it for ~30 days, during which its ID cannot be reused — and unlike custom roles, a create against a soft-deleted ID fails. Prefer `disabled: true` for temporary shutoffs.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
