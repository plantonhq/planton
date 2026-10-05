# GCP Secret Manager Secret - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying a Secret Manager secret using Planton's `GcpSecretManagerSecret` API. The module is written in Go and creates the global trio (`secretmanager.Secret` / `SecretVersion` / `SecretIamMember`) or — when `spec.region` is set — the regional trio (`RegionalSecret` / `RegionalSecretVersion` / `RegionalSecretIamMember`), so one manifest takes a consumer from nothing to a readable, access-granted secret.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Project** with the Secret Manager API enabled (the module enables it if needed)
4. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
5. **IAM permissions**: see [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs

## Directory Structure

```
iac/pulumi/
├── main.go                    # Pulumi program entry point
├── Pulumi.yaml                # Pulumi project configuration
├── README.md                  # This file
└── module/
    ├── main.go                # Module coordinator
    ├── secret.go              # Global/regional secret + version + IAM grants
    ├── locals.go              # Resolved resource + derived values
    └── outputs.go             # Output constants
```

## Quick Start

```bash
cd iac/pulumi
pulumi stack init dev
```

Provide a `iac-input.yaml`:

```yaml
target:
  apiVersion: gcp.planton.dev/v1alpha1
  kind: GcpSecretManagerSecret
  metadata:
    name: db-password
  spec:
    initial_version:
      data:
        value: super-secret-value
    iam_members:
      - role: roles/secretmanager.secretAccessor
        member:
          value: serviceAccount:app@my-project.iam.gserviceaccount.com
```

```bash
pulumi preview
pulumi up
```

## Inputs

The module consumes `GcpSecretManagerSecretIacInput`:

| Field | Required | Description |
|-------|----------|-------------|
| `target` | Yes | `GcpSecretManagerSecret` spec (replication/region, payload, rotation, expiry, IAM grants) |
| `providerConfig` | No | GCP provider configuration; falls back to ambient ADC when omitted |

## Outputs

| Output Key | Type | Description |
|------------|------|-------------|
| `secret_name` | string | Full resource name (`projects/{p}/secrets/{id}` or the regional form) |
| `secret_id` | string | The short secret ID |
| `latest_version_name` | string | `…/versions/1` when `initial_version` was configured; empty otherwise |

## Behavior Notes

- **One kind, two API surfaces**: empty `region` = global secret with replication control; set region = regional secret (payloads never leave the region) with direct CMEK. Scope is permanent.
- **Omitted `replication` renders `auto {}`** — the provider requires the block on global secrets, and automatic placement is the right default when no residency regime applies.
- **The payload is a version, not the secret**: `initial_version.data` seeds version 1 (immutable); later rotations add versions via GCP tooling or pipelines. `initial_version.enabled` is sent explicitly.
- **IAM grants are additive** (`iam_member` semantics) — they compose with grants made elsewhere and never clobber them.
- **Two independent destroy guards**: `deletion_protection` (engine-side plan blocker) and `deletion_policy` (PREVENT/ABANDON/DELETE), evaluated in that order.
- **API enablement**: the module enables `secretmanager.googleapis.com` (with `disable_on_destroy=false`) so a fresh project works first try.

## Related

- [Terraform Module](../tf/README.md) — Terraform implementation
