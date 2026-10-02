# GcpIamDenyPolicy - Terraform Module

This Terraform module provisions an IAM deny policy (`google_iam_deny_policy`). It is the Terraform-side implementation of the Planton `GcpIamDenyPolicy` resource kind and has feature parity with the Pulumi module.

## Overview

A deny policy blocks principals from using specific permissions regardless of any role grants they hold — deny always outranks allow, which makes deny policies the guardrail layer (protect break-glass secrets, forbid destructive APIs org-wide). The policy attaches to a project, folder, or organization; the module renders the URL-encoded full resource name GCP's API expects from the spec's typed parent message, so manifests never hand-assemble it. The deploying principal's permissions are listed in [`../permissions.yaml`](../permissions.yaml); they must be granted at the organization level even for project-attached policies. The module runs on the plain `google` provider — every modeled field is GA on the pinned 8.x line.

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
cd catalog/gcp/gcpiamdenypolicy/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpIamDenyPolicy spec | — |

The `spec` object includes: `parent` (project_id/folder_id/organization_id — at most one; empty falls back to the provider default project via `google_client_config`, count-gated to that one case), `policy_name` (empty defaults to `metadata.name`), `display_name`, `rules` (denied/exception principals and permissions, optional CEL `denial_condition`), and `deletion_policy` (DELETE/PREVENT/ABANDON).

## Outputs

| Name | Description |
|------|-------------|
| `policy_name` | `{url-encoded-parent}/{policy_name}` — the policy's identifier |
| `etag` | The policy's current etag |
