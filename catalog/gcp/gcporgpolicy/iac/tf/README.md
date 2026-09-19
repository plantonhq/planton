# GcpOrgPolicy - Terraform Module

This Terraform module provisions one Google Cloud organization policy (`google_org_policy_policy`): the rules for one constraint at one scope. It is the Terraform-side implementation of the Planton `GcpOrgPolicy` resource kind and has feature parity with the Pulumi module.

## Overview

One resource. The policy's name — `{parent}/policies/{constraint}` — is assembled from the spec's `scope` (a project, folder, or organization; empty means the provider's default project, read once through `data.google_project`) and its constraint (a predefined constraint by literal, or a custom constraint by reference to a `GcpOrgPolicyCustomConstraint`). The enforced rules (`policy`) and the audit-only rules (`dry_run_policy`) render as the provider's `spec` and `dry_run_spec` blocks. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcporgpolicy/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpOrgPolicy spec | — |

The `spec` object includes: `scope` (at most one of `project_id` / `folder_id` / `organization_id`; empty = provider default project), exactly one of `constraint` / `custom_constraint`, `policy` and `dry_run_policy` (each with `inherit_from_parent`, `reset`, and `rules` — every rule carrying exactly one of `allow_all` / `deny_all` / `enforce` / `values`, plus an optional `condition` and `parameters`), and `deletion_policy` (DELETE/PREVENT/ABANDON).

`variables.tf` is generated from the proto contract (`planton tofu generate-variables GcpOrgPolicy`) and formatted with `tofu fmt`; regenerate it when the spec changes rather than editing by hand.

## Outputs

| Name | Description |
|------|-------------|
| `name` | The policy's full resource name, `{parent}/policies/{constraint}` (a project scope is reported as the project number) |
| `etag` | Opaque version marker Google changes on every update |

## Resources Created

- `data.google_project` — read only when the manifest names no scope (count-gated), so every explicitly-scoped plan is credential-free
- `google_org_policy_policy` — the policy, with `spec` and `dry_run_spec` as dynamic blocks emitted exactly when the manifest carries them

## Notes

- **PARITY — rule verdicts.** Google's API models a rule's verdict as booleans in a one-of; the provider flattens that into the strings `"TRUE"` / `"FALSE"` / unset. The spec keeps the API's shape (a oneof of bools) and `locals.tf` renders the string for exactly the arm that is set — so `enforce: false` reaches Google as `"FALSE"` (a real rule) and an unset arm is never sent. The Pulumi module does the same.
- **One policy per constraint per scope.** The name is the identity; a second manifest on the same pair collides. Both halves are immutable — changing either recreates the policy.
- **Rules are set-compared.** The provider suppresses diffs from rule ordering, so re-plans stay clean whatever order the manifest lists rules in.
- **Deletion** removes the policy and the scope falls back to what it inherits. `PREVENT` fails the destroy; `ABANDON` leaves the policy enforced.
