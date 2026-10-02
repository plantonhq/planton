# GcpKmsKeyHandle — Terraform Implementation

This directory contains the Terraform implementation for a Cloud KMS Autokey key handle from the Planton spec: `google_project_service` for the Cloud KMS API and one `google_kms_key_handle`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | The project (provider project when empty) and the handle ID default |
| `main.tf` | `data.google_client_config` (empty project only), `google_project_service`, `google_kms_key_handle` |
| `outputs.tf` | `name`, `kms_key` |

## Send Posture

- **`project`** -- the spec's project, or the provider's project from `google_client_config` (provider configuration, no API call) -- PARITY with the Pulumi module's `GetClientConfig`.
- **`name`** -- `key_handle_name`, defaulting to `metadata.name`.
- **`location`**, **`resource_type_selector`** -- sent as declared.
- **Destroy** -- the provider removes the handle from state only; Google keeps it.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
