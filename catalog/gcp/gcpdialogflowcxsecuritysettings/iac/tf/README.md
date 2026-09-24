# GcpDialogflowCxSecuritySettings — Terraform Implementation

This directory contains the Terraform implementation for Dialogflow CX security settings from the Planton spec: one `google_project_service` (API enablement) and one `google_dialogflow_cx_security_settings`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the display name defaulted from `metadata.name`, null-for-empty optionals |
| `main.tf` | `google_project_service`, `google_dialogflow_cx_security_settings` |
| `outputs.tf` | `name`, `security_settings_id`, `location` |

## Send Posture

- **`display_name`** -- `spec.display_name`, defaulting to `metadata.name` (Google requires one) -- PARITY with the Pulumi module.
- **Redaction, templates, retention strategy** -- sent only when set.
- **`retention_window_days`** -- sent only when positive (0 and unset both mean Google's default TTL).
- **`audio_export_settings`** -- emitted when declared; each leaf sent only when set.
- **`insights_export_settings`** -- the spec's `enable_insights_export`, emitted only when true, so turning it off removes the block.
- **`name` output** -- the resource's `id` (the full path); `security_settings_id` is the provider's `name` (the last segment).

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
