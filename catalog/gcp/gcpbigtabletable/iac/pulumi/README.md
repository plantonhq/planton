# GcpBigtableTable - Pulumi Module

This Pulumi (Go) module provisions a Cloud Bigtable table (`bigtable.Table`) plus one garbage-collection policy (`bigtable.GCPolicy`) per column family that declares one. It is the Pulumi-side implementation of the Planton `GcpBigtableTable` resource kind and has feature parity with the Terraform module.

## Overview

Column families are created on the table with no GC policy; per-family retention lives in the separate GC-policy resources, so a policy change never touches the table object or its data. `deletionProtection` (spec default PROTECTED) is the API-side guard — deletion by ANY client fails until it is set UNPROTECTED. `splitKeys` is ForceNew: changing it REPLACES the table and its data. The module enables the Bigtable Admin API so a fresh project works first try.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpbigtabletable/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — metadata-derived values (table-name fallback)
- `module/table.go` — API enablement + the table + per-family GC policies
- `module/outputs.go` — output keys (must match `outputs.proto`)

## Outputs

| Name | Description |
|------|-------------|
| `table_id` | Fully qualified table resource path |
| `table_name` | Short table name |
| `instance_name` | The parent instance |

## Retention Notes

Bigtable never garbage-collects without a policy — give every family one, or its versions accumulate forever. Expanding what is eligible for collection on a replicated instance requires `ignoreWarnings`.
