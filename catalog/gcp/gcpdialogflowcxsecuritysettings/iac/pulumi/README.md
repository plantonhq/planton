# GcpDialogflowCxSecuritySettings — Pulumi Implementation

This directory contains the Pulumi implementation for Dialogflow CX security settings from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.diagflow.CxSecuritySettings`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `securitySettings` |
| `module/locals.go` | The display name defaulted from `metadata.name` |
| `module/security_settings.go` | Enables the API; maps the settings; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `security_settings_id`, `location`) |

## Send Posture (parity with Terraform)

- **`DisplayName`** -- `spec.display_name`, defaulting to `metadata.name`.
- **Redaction, templates, retention strategy** -- sent only when set.
- **`RetentionWindowDays`** -- sent only when positive.
- **`AudioExportSettings`** -- emitted when declared; each leaf sent only when set.
- **`InsightsExportSettings`** -- emitted only when `enable_insights_export` is true.
- **`name` output** -- the resource's ID (the full path).

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
