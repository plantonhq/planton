# GcpGcsBucketIamMember - Terraform Module

This Terraform module provisions a single additive bucket-scoped IAM grant (`google_storage_bucket_iam_member`). It is the Terraform-side implementation of the Planton `GcpGcsBucketIamMember` resource kind and has feature parity with the Pulumi module.

## Overview

The module merges one (role, member[, condition]) pair into the target bucket's IAM policy without touching any other member's bindings; destroy subtracts only this exact pair. Every argument is immutable (ForceNew) — IAM grants have no update, so any change replaces the grant. There is no project input: bucket names are global.

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
cd catalog/gcp/gcpgcsbucketiammember/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpGcsBucketIamMember spec | — |

The `spec` object includes: `bucket` (the bucket name, not a `gs://` URL), `role` (predefined or full custom role name), `member` (IAM member format), and an optional `condition` (`title`, `expression`, optional `description`) rendered as a `condition` block only when set.

## Outputs

| Name | Description |
|------|-------------|
| `bucket` | The bucket whose policy received the grant |
| `role` | The granted role |
| `member` | The granted member |
| `etag` | The bucket IAM policy etag after the grant |
