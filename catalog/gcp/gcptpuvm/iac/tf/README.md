# GcpTpuVm — Terraform Implementation

This directory contains the Terraform implementation for a Cloud TPU VM from the Planton spec: one `google_project_service` (API enablement, GA provider) and one `google_tpu_v2_vm` on `google-beta`, attached under the recorded admission in `pkg/providerparity/admissions/google-beta.yaml`.

## Provider

- **Providers**: `hashicorp/google` and `hashicorp/google-beta`, both `~> 8.3` (one release line)
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block with both providers and why the beta channel is attached |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the node id defaulted from `metadata.name`, null-for-empty optionals, network interfaces normalized, the `planton-ai_*` labels |
| `main.tf` | `google_project_service`, `google_tpu_v2_vm` (`provider = google-beta`) |
| `outputs.tf` | `name`, `node_id`, `zone` |

## Send Posture

- **`name`** -- `spec.node_id`, defaulting to `metadata.name` -- PARITY with the Pulumi module.
- **`accelerator_type`** -- sent only when set; with no accelerator declared the provider asks for a `v2-8`.
- **`network_config` vs `network_configs`** -- one interface renders the singular block, several the repeated one.
- **Compute references** -- subnetwork and data-disk self-links trimmed to relative paths.
- **`shielded_instance_config`** -- sent only when `enable_secure_boot` is true.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
