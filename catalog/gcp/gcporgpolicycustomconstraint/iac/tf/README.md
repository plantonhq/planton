# GcpOrgPolicyCustomConstraint - Terraform Module

This Terraform module provisions one custom organization-policy constraint (`google_org_policy_custom_constraint`): a rule the organization writes itself over a Google Cloud resource's fields. It is the Terraform-side implementation of the Planton `GcpOrgPolicyCustomConstraint` resource kind and has feature parity with the Pulumi module.

## Overview

One resource. The constraint's name is `custom.{constraint_name}` — the spec takes the bare name (defaulting to `metadata.name`) and the module adds Google's `custom.` prefix, so it can never be doubled or forgotten. The parent is always the organization. The constraint is a DEFINITION: nothing is enforced until a `GcpOrgPolicy` references its `constraint` output. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcporgpolicycustomconstraint/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpOrgPolicyCustomConstraint spec | — |

The `spec` object includes: `organization_id`, `constraint_name` (bare, without `custom.`; empty defaults to `metadata.name`), `display_name`, `description` (the violation message users see), `resource_types` (immutable), `method_types`, `condition` (CEL over the resource), `action_type` (ALLOW/DENY), and `deletion_policy` (DELETE/PREVENT/ABANDON).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpOrgPolicyCustomConstraint`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `name` | `organizations/{org}/customConstraints/custom.{name}` — the full resource name (the resource id) |
| `constraint` | `custom.{name}` — the handle a `GcpOrgPolicy`'s `custom_constraint` references |
| `update_time` | RFC 3339 last-update timestamp |

## Resources Created

- `google_org_policy_custom_constraint` — the constraint, with optional strings and `deletion_policy` sent only when set

## Notes

- **Immutable identity.** The name, the organization, and `resource_types` recreate the constraint when changed — and every policy enforcing the old name lapses. The condition, action, methods, display name, and description update in place.
- **Destroy the policies first.** Deleting a constraint that policies still enforce makes those policies fail to apply; a chart whose policies reference this resource destroys them in the right order.
- **Supported services and fields** are Google's list; a constraint over an unsupported resource type or method is rejected at apply time.
