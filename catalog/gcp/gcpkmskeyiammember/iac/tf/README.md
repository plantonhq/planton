# GcpKmsKeyIamMember - Terraform Module

This Terraform module provisions a single additive key-scoped IAM grant (`google_kms_crypto_key_iam_member`). It is the Terraform-side implementation of the Planton `GcpKmsKeyIamMember` resource kind and has feature parity with the Pulumi module.

## Overview

The module merges one (role, member[, condition]) pair into the target crypto key's IAM policy without touching any other member's bindings; destroy subtracts only this exact pair. Every argument is immutable (ForceNew) — IAM grants have no update, so any change replaces the grant atomically. There is no project or location input: both are embedded in the key's resource path.

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
cd catalog/gcp/gcpkmskeyiammember/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpKmsKeyIamMember spec | — |

The `spec` object includes: `crypto_key_id` (fully-qualified key path, validated), `role` (predefined or fully-qualified custom role name), `member` (IAM member format, validated — deleted principals rejected), optional `condition` (`title`, `expression`, optional `description`).

## Outputs

| Name | Description |
|------|-------------|
| `crypto_key_id` | The key whose policy received the grant |
| `role` | The granted role |
| `member` | The granted member |
| `etag` | The key IAM policy etag after the grant |
