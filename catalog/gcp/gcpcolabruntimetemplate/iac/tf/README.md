# GcpColabRuntimeTemplate — Terraform Implementation

This directory contains the Terraform implementation for a Colab Enterprise runtime template from the Planton spec: one `google_project_service` (API enablement) and one `google_colab_runtime_template`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the template id and display name defaulted from `metadata.name`, the subnetwork trimmed to its relative path, the `planton-ai_*` labels |
| `main.tf` | `google_project_service`, `terraform_data.location` (the location replace trigger), `google_colab_runtime_template` |
| `outputs.tf` | `name`, `runtime_template_id`, `location` |

## Send Posture

- **`name`** / **`display_name`** -- `spec.runtime_template_id` / `spec.display_name`, each defaulting to `metadata.name` -- PARITY with the Pulumi module.
- **Optional+Computed settings** (machine type, accelerator count, disk type and size, idle timeout, EUC, Secure Boot, image release) -- sent only when set.
- **`disk_size_gb`** -- the spec's `int64` as a decimal string.
- **Lifted wrappers** -- `idle_shutdown_config`, `euc_config`, `shielded_vm_config`, `encryption_spec` blocks rendered only when their spec field is set.
- **`subnetwork`** -- a `GcpSubnetwork` self-link trimmed to its relative path.
- **`location`** -- the provider plans a changed location as an in-place update at the old path; `terraform_data.location` and `replace_triggered_by` make it a replacement -- PARITY with the Pulumi module's `ReplaceOnChanges`.
- **Fixed at creation** -- the provider never sends a changed `idle_timeout`, and turning `euc_disabled` or `enable_secure_boot` off sends nothing; a different value needs a new template.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
