# GcpBinaryAuthorizationPolicy — Terraform Implementation

This directory contains the Terraform implementation for a project's Binary Authorization policy from the Planton spec: `google_project_service` for the Binary Authorization API and one `google_binary_authorization_policy`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | The project (provider project when empty), send-only-when-set values |
| `main.tf` | `data.google_client_config` (empty project only), `google_project_service`, `google_binary_authorization_policy` |
| `outputs.tf` | `name`, `project_id` |

## Send Posture

- **`project`** -- the spec's project, or the provider's project from `google_client_config` (provider configuration, no API call) -- PARITY with the Pulumi module's `GetClientConfig`.
- **`admission_whitelist_patterns`** -- one block per spec string.
- **`require_attestations_by`** -- sent only when non-empty.
- **`description`**, **`global_policy_evaluation_mode`**, **`deletion_policy`** -- sent only when set.
- **Destroy** -- under `DELETE` the provider writes Google's default policy back.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
