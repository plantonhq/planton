# GcpSecretManagerSecret - Terraform Module

This Terraform module provisions a Secret Manager secret with its optional first version and secret-scoped IAM grants (`google_secret_manager_secret` / `_secret_version` / `_secret_iam_member`, or their `_regional_` variants when `spec.region` is set). It is the Terraform-side implementation of the Planton `GcpSecretManagerSecret` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates the secret container, optionally seeds version 1 from `initial_version`, and applies additive IAM grants — so one manifest takes a consumer from nothing to a readable, access-granted secret.

One kind, two API surfaces: an empty `region` creates a global secret with replication control (an omitted `replication` renders the API's `auto {}` mode); a set region creates a regional secret whose payloads never leave that region, with CMEK attached directly. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpsecretmanagersecret/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpSecretManagerSecret spec | — |

The `spec` object includes: `secret_id` (empty defaults to `metadata.name`), `region` (empty means global), `replication` (global only — auto with optional CMEK, or user_managed replicas), `customer_managed_encryption` (regional only), `initial_version` (data/enabled/is_base64/deletion_policy — seeds version 1), `iam_members` (additive role+member grants with optional conditions), `expire_time` XOR `ttl`, `version_destroy_ttl` (delayed destruction), `rotation` + `topics` (Pub/Sub rotation reminders), `version_aliases`, `annotations`, `tags`, `deletion_protection`, `project_id` (empty falls back to the provider default project), `labels`, and `deletion_policy` (DELETE/PREVENT/ABANDON).

## Outputs

| Name | Description |
|------|-------------|
| `secret_name` | Full resource name (global or regional form) |
| `secret_id` | The short secret ID |
| `latest_version_name` | `…/versions/1` when `initial_version` was configured; empty otherwise |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs.
