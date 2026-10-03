# GcpBigtableTable - Terraform Module

This Terraform module provisions a Cloud Bigtable table (`google_bigtable_table`) plus one garbage-collection policy (`google_bigtable_gc_policy`) per column family that declares one. It is the Terraform-side implementation of the Planton `GcpBigtableTable` resource kind and has feature parity with the Pulumi module.

## Overview

Column families are created on the table with no GC policy; per-family retention lives in the separate GC-policy resources, so a policy change never touches the table object or its data. `deletion_protection` (spec default PROTECTED) is the API-side guard — deletion by ANY client fails until it is set UNPROTECTED. `split_keys` is ForceNew: changing it REPLACES the table and its data. The module enables the Bigtable Admin API so a fresh project works first try, and runs on the plain `google` provider — every modeled field is GA at the pinned release.

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
cd catalog/gcp/gcpbigtabletable/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpBigtableTable spec | — |

The `spec` object includes: `instance` (parent instance short name; required, ForceNew), `table_name` (empty falls back to `metadata.name`; ForceNew), `column_families` (each with optional `gc_policy`: typed `mode`/`max_age`/`max_versions` XOR raw `gc_rules` JSON, plus `ignore_warnings`), `split_keys` (ForceNew), `change_stream_retention`, `automated_backup_policy`, `deletion_protection` (default PROTECTED), `row_key_schema`, and `project_id` (empty falls back to the provider default project).

## Outputs

| Name | Description |
|------|-------------|
| `table_id` | Fully qualified table resource path |
| `table_name` | Short table name |
| `instance_name` | The parent instance |

## Retention Notes

Bigtable never garbage-collects without a policy — give every family one, or its versions accumulate forever. Expanding what is eligible for collection on a replicated instance requires `ignore_warnings`.
