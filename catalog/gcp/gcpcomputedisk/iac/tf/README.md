# GcpComputeDisk - Terraform Module

This Terraform/OpenTofu module provisions a zonal Compute Engine persistent disk (`google_compute_disk`). It is the Terraform-side implementation of the Planton `GcpComputeDisk` resource kind and has feature parity with the Pulumi module.

## Overview

The module enables the Compute Engine API (`disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. `disk_name` falls back to `metadata.name`; user labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Pulumi module. Empty optional strings are normalized to `null` so the provider applies its own defaults (default disk type, Google-managed encryption, `READ_WRITE_SINGLE` access) instead of receiving empty strings it would reject.

**Immutability is the sharp edge**: name, zone, type, sources, encryption, and architecture are ForceNew — changing them replaces the disk and its data; size grows in place but never shrinks. At most one source (image / snapshot / instant snapshot / storage object / source_disk) is enforced pre-deploy by the spec's CEL; none creates an empty disk. Deleting a disk still attached to a running instance fails; `create_snapshot_before_destroy` takes a final snapshot during destroy (CMEK disks reuse their key for it), and `deletion_policy: PREVENT` fails the destroy outright. For CMEK, the Compute Engine service agent must hold `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key before create; encrypted source images/snapshots decrypt via `source_image_encryption` / `source_snapshot_encryption` (CMEK only — raw CSEK keys are deliberately not part of the contract). `async_primary_disk` pairs the disk as an async-replication secondary; `guest_os_features` and `licenses` cover bootable-disk and BYOL-import shapes.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu apply --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Module Layout

- `provider.tf` — google provider pin
- `variables.tf` — the converter-contract `metadata`/`spec` variables
- `locals.tf` — disk-name fallback, empty-string→null normalization, label merge
- `main.tf` — API enablement + the disk (CMEK and resource-manager-tag blocks emitted only when set)
- `outputs.tf` — `name`, `disk_id`, `self_link`, `zone`, `size_gb`, `type`

## Outputs

| Output | Description |
|--------|-------------|
| `name` | Name of the disk in GCP |
| `disk_id` | Server-assigned unique numeric identifier |
| `self_link` | Self-link URL — the attachment composition key |
| `zone` | Zone the disk lives in (plain zone name) |
| `size_gb` | Provisioned size in GB |
| `type` | Disk type, normalized to the plain type name (identical on both engines) |
