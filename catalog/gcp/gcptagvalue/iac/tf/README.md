# GcpTagValue - Terraform Module

This Terraform module provisions one Resource Manager tag value (`google_tags_tag_value`). It is the Terraform-side implementation of the Planton `GcpTagValue` resource kind and has feature parity with the Pulumi module.

## Overview

One resource. The value's key is the spec's `tag_key` (a reference to a `GcpTagKey`'s `name` output, `tagKeys/{id}`), its short name defaults to `metadata.name`, and the optional description is sent only when set. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcptagvalue/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpTagValue spec | — |

The `spec` object includes: `tag_key` (`tagKeys/{id}`), `short_name` (empty defaults to `metadata.name`), `description`, and `deletion_policy` (DELETE/PREVENT/ABANDON).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpTagValue`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `name` | `tagValues/{id}` — what a `GcpTagBinding`'s `tag_value` references |
| `namespaced_name` | `{parent}/{key_short_name}/{short_name}` |
| `tag_value_id` | The bare numeric id |
| `create_time` | RFC 3339 creation timestamp |

## Resources Created

- `google_tags_tag_value` — the value, with the description and `deletion_policy` sent only when set

## Notes

- **The key and the short name are immutable**; only the description updates in place.
- **A value with bindings cannot be deleted** — destroy the `GcpTagBinding`s first (a chart's dependency order does this when they reference the value).
- **Deletion is soft.** A deleted value's short name stays reserved under the same key for 30 days.
