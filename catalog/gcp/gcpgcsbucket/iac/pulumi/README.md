# GcpGcsBucket - Pulumi Module

This Pulumi (Go) module provisions a Cloud Storage bucket (`storage.Bucket`) plus additive per-bucket IAM grants and the bucket's structural companions — folders (`storage.Folder`), managed folders (`storage.ManagedFolder`), and Pub/Sub notification configs (`storage.Notification`). It is the Pulumi-side implementation of the Planton `GcpGcsBucket` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Cloud Storage API (`disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. User labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Terraform module.

**Safety is the sharp edge**: `force_destroy` honors the spec (default false — destroying a non-empty bucket fails instead of erasing data); `deletion_policy: PREVENT` fails the destroy outright (ABANDON unmanages without deleting); a locked retention policy is irreversible and blocks deletion until objects pass retention; the soft-delete block is sent only when the spec sets it, so unset specs follow GCP's server-side 7-day default without a perpetual diff. Numeric lifecycle conditions ride on presence — a set `0` is sent via the provider's send-zero flags, identically to the Terraform module — and size bands (`size_above_bytes` / `size_below_bytes`) follow the same presence contract. One `Encryption` block carries both the default CMEK key and the per-encryption-type enforcement for new objects (`NotRestricted` / `FullyRestricted` per GMEK/CMEK/CSEK).

**Structural companions**: folders are created parents-before-children — the module sorts by depth and makes each folder depend on its nearest managed ancestor, because the Storage API never auto-creates missing parents (destroy runs the reverse order). Each folder carries its own `force_destroy` (a client-side object sweep), and the kind-level `deletion_policy` guards folders and managed folders exactly like the bucket. Managed folders are independent prefix anchors; their `force_destroy` is server-side (`allowNonEmpty`) and non-destructive to data. Notification configs are immutable end to end (every change replaces) and are named by list index — matching the Terraform module's keys; the topic must be the fully-qualified `projects/{project}/topics/{name}` form, and the project's GCS service agent must already hold `roles/pubsub.publisher` on it (a composed grant; see the kind README).

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Module Layout

- `main.go` — Pulumi entrypoint (loads the IaC input, calls the module)
- `module/main.go` — provider setup + orchestration
- `module/locals.go` — label merge
- `module/gcs_bucket.go` — API enablement + the bucket + additive IAM members + outputs
- `module/bucket_companions.go` — folders (depth-ordered) + managed folders + notification configs
- `module/outputs.go` — output key constants

## Outputs

| Output | Description |
|--------|-------------|
| `bucket_id` | Bucket ID (equals the bucket name) |
| `bucket_name` | Bucket name |
| `url` | `gs://<name>` |
| `self_link` | API self link |
| `location` | Location as reported by GCS |
| `project_number` | Numeric owning project |
