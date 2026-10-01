# GcpComputeImage — Terraform Implementation

This directory contains the Terraform implementation for a Compute Engine custom image from the Planton spec: `google_project_service` for the Compute Engine API and one `google_compute_image`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Project, name default, optional strings and lists as null, attribution labels, the source keys |
| `main.tf` | `google_project_service`, `google_compute_image` |
| `outputs.tf` | `name`, `self_link`, `family`, `disk_size_gb` |

## Send Posture

- **`name`** -- `image_name`, defaulting to `metadata.name`.
- **Optional+Computed** -- `disk_size_gb`, `guest_os_features`, `licenses`, `storage_locations`, and `shielded_instance_initial_state` sent only when declared, so values Google inherits from the source never show as a diff.
- **Keys** -- `kms_key` becomes `image_encryption_key.kms_key_self_link`; each `source*_encryption` block its `*_encryption_key` block; raw keys are never sent.
- **`family` output** -- empty when the image has none, on both engines.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
