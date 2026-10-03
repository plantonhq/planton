# GcpCloudSqlDatabase - Terraform Module

This Terraform module provisions a logical database (`google_sql_database`) inside a Cloud SQL instance. It is the Terraform-side implementation of the Planton `GcpCloudSqlDatabase` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates one database on an existing instance, referenced by name. Charset and collation are engine-specific and validated by the API at deploy time. No API enablement: the hosting instance cannot exist without `sqladmin.googleapis.com` already enabled.

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
cd catalog/gcp/gcpcloudsqldatabase/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `spec` | GcpCloudSqlDatabase spec (project_id, instance, database_name, charset, collation) | — |
| `metadata` | Resource metadata (name, org, env, labels) | — |

`StringValueOrRef` fields (`project_id`, `instance`) arrive as plain strings after the CLI's ref resolution.

## Outputs

| Name | Description |
|------|-------------|
| `database_name` | Name of the database inside the instance |
| `self_link` | Self-link URL of the database resource |
