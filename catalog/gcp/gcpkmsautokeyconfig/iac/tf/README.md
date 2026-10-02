# GcpKmsAutokeyConfig — Terraform Implementation

This directory contains the Terraform implementation for a Cloud KMS Autokey configuration from the Planton spec: `google_project_service` for the Cloud KMS API, and one `google_kms_autokey_config` (folder) or `google_kms_project_autokey_config` (project).

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Scope selection, the bare folder ID, the project and key project, the API projects |
| `main.tf` | `data.google_client_config` (empty scope only), `google_project_service` per API project, the folder or project configuration |
| `outputs.tf` | `name`, `parent` |

## Send Posture

- **Scope** -- a folder arm creates the folder configuration with Google's bare numeric folder ID; otherwise the project configuration, on the project arm's project or, for an empty scope, the provider's project from `google_client_config` (provider configuration, no API call) -- PARITY with the Pulumi module's `GetClientConfig`.
- **`key_project`** -- sent as `projects/{id}` when set (folder only).
- **`key_project_resolution_mode`**, **`deletion_policy`** -- sent only when set.
- **APIs** -- `cloudkms.googleapis.com` on the project configuration's project and on the key project; `disable_on_destroy = false`.
- **Destroy** -- under `DELETE` the provider PATCHes the configuration empty.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
