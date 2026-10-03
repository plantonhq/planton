# GcpFirestoreIndex - Pulumi Module

This Pulumi (Go) module provisions a Firestore composite index (`firestore.Index`). It is the Pulumi-side implementation of the Planton `GcpFirestoreIndex` resource kind and has feature parity with the Terraform module.

## Overview

Every index property is immutable — changing anything replaces the index (Firestore rebuilds in the background). Fields map to order, array_config, vector_config (with flat index), or the Enterprise search_config (text/geo). `multikey` and `unique` cover the MongoDB-compatible scope, `skip_wait` enables fire-and-forget creation, and `deletion_policy` guards destroys (PREVENT fails; ABANDON unmanages without deleting). The module enables the Firestore API so a fresh project works first try.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — resolved IaC input values
- `module/index.go` — API enablement + the composite index
- `module/outputs.go` — output keys (must match `outputs.proto`)

## Outputs

| Name | Description |
|------|-------------|
| `index_id` | Fully qualified index resource name |
| `collection` | Collection (group) the index applies to |
