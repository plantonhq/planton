# GcpFolder - Terraform Module

This Terraform module provisions a Google Cloud Resource Manager folder (`google_folder`). It is the Terraform-side implementation of the Planton `GcpFolder` resource kind and has feature parity with the Pulumi module.

## Overview

One resource. The folder's parent is rendered from the spec's `parent` message (`organizations/{id}` for a top-level folder, `folders/{id}` for a nested one — the nested form is a reference to another `GcpFolder`), its display name defaults to `metadata.name`, and the destroy guard (`deletion_protection`) is always sent explicitly so the spec is the single source of truth. Changing the parent MOVES the folder in place; only the create-time `tags` are immutable. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcpfolder/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpFolder spec | — |

The `spec` object includes: `parent` (exactly one of `organization_id` / `folder_id`), `display_name` (empty defaults to `metadata.name`), `deletion_protection` (null defaults to true), `tags` (create-time only, `tagKeys/{id}` → `tagValues/{id}`), and `deletion_policy` (DELETE/PREVENT/ABANDON).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpFolder`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `folder_id` | The folder's numeric ID — what every child references (a nested folder's parent, a project's `folder_id`, a policy's or binding's folder scope) |
| `name` | `folders/{folder_id}` — the resource name Google's APIs and IAM use |
| `lifecycle_state` | `ACTIVE`, or `DELETE_REQUESTED` during the 30-day soft-delete window |
| `create_time` | RFC 3339 creation timestamp |

## Resources Created

- `google_folder` — the folder, with `deletion_protection` always sent and `tags` / `deletion_policy` sent only when set

## Notes

- **The destroy guard is on by default.** With `deletion_protection` unset or true, a destroy fails until the field is set to false and applied first. `deletion_policy = ABANDON` is evaluated before the guard, so abandoning a protected folder still works; `PREVENT` is evaluated before both.
- **A non-empty folder cannot be deleted.** Google refuses while any project or folder is inside it — destroy the children first (a chart's dependency order does this when they reference the folder).
- **Deletion is soft.** A deleted folder is recoverable for 30 days, and its display name stays reserved under the same parent for that window.
- **Tags at create time recreate the folder when changed.** Bind tags to an existing folder with `GcpTagBinding` instead.
