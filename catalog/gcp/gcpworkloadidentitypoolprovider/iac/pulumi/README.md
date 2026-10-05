# GCP Workload Identity Pool Provider - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP Workload Identity Pool Providers using Planton's `GcpWorkloadIdentityPoolProvider` API. The module is written in Go and uses the Pulumi GCP provider to create `iam.WorkloadIdentityPoolProvider` resources (backed by `google_iam_workload_identity_pool_provider`).

A provider attaches one external issuer (OIDC, AWS, SAML, or X.509) to a Workload Identity Pool — the piece that makes keyless federation work. Its `name` output is the audience tokens are minted for.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **An existing Workload Identity Pool** — the provider attaches to it by pool ID
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
    ├── main.go                              # Module coordinator
    ├── workload_identity_pool_provider.go   # Provider resource creation
    ├── locals.go                            # Resolved resource holder
    └── outputs.go                           # Output constants
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Create Input File

Provide a `iac-input.yaml` with the provider specification (GitHub Actions example):

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpWorkloadIdentityPoolProvider
  metadata:
    name: github-oidc
  spec:
    workload_identity_pool_id:
      value: github-actions
    workload_identity_pool_provider_id: github-oidc
    attribute_mapping:
      google.subject: assertion.sub
      attribute.repository: assertion.repository
    attribute_condition: assertion.repository_owner == "my-org"
    oidc:
      issuer_uri: https://token.actions.githubusercontent.com
```

### 3. Build, Preview, and Deploy

```bash
make build
pulumi preview
pulumi up
```

### 4. View Outputs

```bash
pulumi stack output name   # the audience string for token exchange
```

## Inputs

The module consumes `GcpWorkloadIdentityPoolProviderIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpWorkloadIdentityPoolProvider` spec (pool ref, provider ID, issuer arm, mappings) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

Spec fields: `workload_identity_pool_id` (required ref, immutable), `workload_identity_pool_provider_id` (required, immutable), `project_id` (falls back to the provider default project when empty), optional `display_name`, `description`, `disabled`, `attribute_mapping` (required for OIDC — must include `google.subject`), `attribute_condition`, and exactly one issuer arm (`aws` | `oidc` | `saml` | `x509`).

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `name` | string | Full provider resource name — **the audience for token exchange** |
| `workload_identity_pool_provider_id` | string | The bare provider ID |
| `state` | string | `ACTIVE`, or `DELETED` while soft-deleted |

## Behavior Notes

- **Immutability**: pool, provider ID, and project are ForceNew, and the issuer TYPE cannot change on a live provider; the chosen arm's contents, mappings, and condition update in place.
- **Soft delete without undelete**: destroying the provider soft-deletes it for ~30 days, during which its ID cannot be reused; a create against a soft-deleted ID fails. Prefer `disabled: true`.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
