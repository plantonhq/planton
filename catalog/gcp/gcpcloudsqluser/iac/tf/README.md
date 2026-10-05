# GcpCloudSqlUser - Terraform Module

This Terraform module provisions a database user (`google_sql_user`) on a Cloud SQL instance. It is the Terraform-side implementation of the Planton `GcpCloudSqlUser` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates one user on an existing instance: a classic BUILT_IN username/password credential (with an optional per-user password policy) or a passwordless IAM-authenticated principal (user, service account, or group). The password flows from a `(sensitive)`-annotated spec field and is never exported in outputs. No API enablement: the hosting instance cannot exist without `sqladmin.googleapis.com` already enabled.

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
cd catalog/gcp/gcpcloudsqluser/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `spec` | GcpCloudSqlUser spec (project_id, instance, user_name, password, type, host, password_policy) | — |
| `metadata` | Resource metadata (name, org, env, labels) | — |

`StringValueOrRef` fields (`project_id`, `instance`) arrive as plain strings after the CLI's ref resolution.

## Outputs

| Name | Description |
|------|-------------|
| `user_name` | The user name as stored by Cloud SQL |
