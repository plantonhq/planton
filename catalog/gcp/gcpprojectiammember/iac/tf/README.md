# GcpProjectIamMember - Terraform Module

This Terraform module provisions a single additive project-level IAM grant (`google_project_iam_member`). It is the Terraform-side implementation of the Planton `GcpProjectIamMember` resource kind and has feature parity with the Pulumi module.

## Overview

The module merges one (role, member[, condition]) pair into the target project's IAM policy without touching any other member's bindings; destroy subtracts only this exact pair. Every argument is immutable (ForceNew) — IAM grants have no update, so any change replaces the grant atomically. When `project_id` is empty, the module resolves the provider's default project via the `google_client_config` data source (the underlying resource requires an explicit project).

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
cd catalog/gcp/gcpprojectiammember/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpProjectIamMember spec | — |

The `spec` object includes: `role` (object with `value`; predefined or fully-qualified custom role name), `member` (object with `value`; IAM member format, validated — deleted principals rejected), `project_id` (object with `value`; empty falls back to the provider default project), optional `condition` (`title`, `expression`, optional `description`).

## Outputs

| Name | Description |
|------|-------------|
| `project_id` | The project whose policy received the grant |
| `role` | The granted role |
| `member` | The granted member |
| `etag` | The project IAM policy etag after the grant |
