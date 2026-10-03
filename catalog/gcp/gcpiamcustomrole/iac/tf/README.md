# GcpIamCustomRole - Terraform Module

This Terraform module provisions a project-scoped GCP IAM custom role (`google_project_iam_custom_role`). It is the Terraform-side implementation of the Planton `GcpIamCustomRole` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates a named, least-privilege permission bundle grantable as `projects/<project>/roles/<role_id>`. `role_id` and `project` are immutable (ForceNew); `title`, `description`, `stage`, and `permissions` update in place, and permission edits propagate immediately to every existing grant of the role. GCP soft-deletes custom roles for up to 14 days; re-creating a role with a soft-deleted ID undeletes and patches it — the provider handles this natively.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpiamcustomrole/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpIamCustomRole spec | — |

The `spec` object includes: `role_id` (3-64 chars, no hyphens), `project_id` (object with `value`; empty falls back to the provider default project), `title`, `description` (optional), `permissions` (min 1), `stage` (optional, default `GA`).

## Outputs

| Name | Description |
|------|-------------|
| `name` | Fully-qualified role name (`projects/<project>/roles/<role_id>`) — the grantable handle |
| `role_id` | The bare role ID within the project |
| `deleted` | Whether the role is currently soft-deleted |
