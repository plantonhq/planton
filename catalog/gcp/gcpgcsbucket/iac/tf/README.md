# GcpGcsBucket - Terraform Module

This Terraform/OpenTofu module provisions a Cloud Storage bucket (`google_storage_bucket`) plus additive per-bucket IAM grants and the bucket's structural companions — folders (`google_storage_folder`), managed folders (`google_storage_managed_folder`), and Pub/Sub notification configs (`google_storage_notification`). It is the Terraform-side implementation of the Planton `GcpGcsBucket` resource kind and has feature parity with the Pulumi module.

## Overview

The module enables the Cloud Storage API (`disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. User labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Pulumi module.

**Safety is the sharp edge**: `force_destroy` honors the spec (default false — destroying a non-empty bucket fails instead of erasing data); `deletion_policy: PREVENT` fails the destroy outright (ABANDON unmanages without deleting); a locked retention policy is irreversible and blocks deletion until objects pass retention; the soft-delete block is sent only when the spec sets it, so unset specs follow GCP's server-side 7-day default without a perpetual diff. Numeric lifecycle conditions ride on presence — a set `0` is sent via the provider's send-zero flags, identically to the Pulumi module — and size bands (`size_above_bytes` / `size_below_bytes`) follow the same presence contract. One `encryption` block carries both the default CMEK key and the per-encryption-type enforcement for new objects (`NotRestricted` / `FullyRestricted` per GMEK/CMEK/CSEK).

**Structural companions**: folders are created in depth-tiered resource groups (`folders_l1`..`folders_l5`) chained with `depends_on`, because the Storage API never auto-creates missing parents and instances of one `for_each` cannot order among themselves — parents create first, children destroy first. Each folder carries its own `force_destroy` (a client-side object sweep), and the kind-level `deletion_policy` guards folders and managed folders exactly like the bucket. Managed folders are independent prefix anchors (`for_each` keyed by path); their `force_destroy` is server-side (`allowNonEmpty`) and non-destructive to data. Notification configs are immutable end to end (every change replaces) and are keyed by list index; the module passes the topic through as-is — it must be the fully-qualified `projects/{project}/topics/{name}` form, and the project's GCS service agent must already hold `roles/pubsub.publisher` on it (a composed grant; see the kind README).

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu apply --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Module Layout

- `provider.tf` — google provider pin (`~> 8.3`; all fields GA on the released line)
- `variables.tf` — the converter-contract `metadata`/`spec` variables
- `locals.tf` — ambient-project fallback, empty-string→null normalization, label merge, IAM grant keying, folder depth split, notification index keying
- `main.tf` — API enablement + the bucket + additive IAM members + folders / managed folders / notification configs
- `outputs.tf` — `bucket_id`, `bucket_name`, `url`, `self_link`, `location`, `project_number`

## Outputs

| Output | Description |
|--------|-------------|
| `bucket_id` | Bucket ID (equals the bucket name) |
| `bucket_name` | Bucket name |
| `url` | `gs://<name>` |
| `self_link` | API self link |
| `location` | Location as reported by GCS |
| `project_number` | Numeric owning project |
