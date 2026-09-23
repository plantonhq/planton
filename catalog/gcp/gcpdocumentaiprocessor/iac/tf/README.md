# GcpDocumentAiProcessor — Terraform Implementation

This directory contains the Terraform implementation for a Document AI processor from the Planton spec: one `google_project_service` (API enablement), one `google_document_ai_processor`, and an optional `google_document_ai_processor_default_version`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the display name defaulted from `metadata.name`, null-for-empty optionals |
| `main.tf` | `google_project_service`, `google_document_ai_processor`, `google_document_ai_processor_default_version` (count-gated on `default_version`) |
| `outputs.tf` | `name`, `processor_id`, `location`, `process_endpoint` |

## Send Posture

- **`display_name`** -- `spec.display_name`, defaulting to `metadata.name` -- PARITY with the Pulumi module.
- **`kms_key_name`**, **`deletion_policy`** -- sent only when set.
- **`version`** -- `{processor id}/processorVersions/{default_version}`, composed the same way as the Pulumi module; the binding exists only when `default_version` is set.
- **`name` output** -- the Terraform `id` (the full path); `processor_id` is the provider's short `name`.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
