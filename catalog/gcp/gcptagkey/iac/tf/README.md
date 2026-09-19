# GcpTagKey - Terraform Module

This Terraform module provisions one Resource Manager tag key (`google_tags_tag_key`). It is the Terraform-side implementation of the Planton `GcpTagKey` resource kind and has feature parity with the Pulumi module.

## Overview

One resource. The key's owner is rendered from the spec's `parent` message (`organizations/{id}` or `projects/{id}` — a folder cannot own a key), its short name defaults to `metadata.name`, and the optional description, purpose, purpose data, and allowed-values regex are sent only when set. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcptagkey/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpTagKey spec | — |

The `spec` object includes: `parent` (exactly one of `organization_id` / `project_id`), `short_name` (empty defaults to `metadata.name`), `description`, `purpose` (GCE_FIREWALL / DATA_GOVERNANCE), `purpose_data`, `allowed_values_regex`, and `deletion_policy` (DELETE/PREVENT/ABANDON).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpTagKey`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `name` | `tagKeys/{id}` — what a `GcpTagValue`'s `tag_key` references |
| `namespaced_name` | `{org_or_project}/{short_name}` — the form `resource.matchTag` tests |
| `tag_key_id` | The bare numeric id |
| `create_time` | RFC 3339 creation timestamp |

## Resources Created

- `google_tags_tag_key` — the key, with every optional input and `deletion_policy` sent only when set

## Notes

- **Almost everything is immutable.** The owner, short name, purpose, and purpose data recreate the key when changed — and Google refuses to delete a key that still has values. Only the description and the allowed-values regex update in place.
- **Deletion is soft.** A deleted key's short name stays reserved under the same owner for 30 days.
- **A folder cannot own a key** — Google's rule; the spec's parent has no folder arm.
