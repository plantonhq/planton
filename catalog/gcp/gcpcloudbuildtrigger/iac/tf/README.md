# GcpCloudBuildTrigger — Terraform Implementation

This directory contains the Terraform implementation for a Cloud Build trigger from the Planton spec: `google_project_service` for the Cloud Build API and one `google_cloudbuild_trigger`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Project, trigger name default, location, deletion policy |
| `main.tf` | `google_project_service` and the trigger, with one dynamic block per event source and build block |
| `outputs.tf` | `id`, `trigger_id`, `name` |

## Send Posture

- **Name** -- `spec.triggerName`, or `metadata.name` when empty; renames in place.
- **Location** -- not sent when empty, so the provider's `global` applies.
- **Optional arguments** -- sent only when set (empty strings, zero numbers, empty lists and maps, and `false` booleans become `null`), so Google's defaults apply; the build `timeout` falls to the provider's `600s`.
- **Build lists** -- `steps` and `secrets` render as the provider's `step` and `secret` blocks, in spec order.
- **Never sent** -- `build.options.dynamic_substitutions` and `substitution_option` (Google fixes both for triggered builds) and `build.step.timing` (output only).
- **No labels** -- a trigger has `tags`, not labels, so no attribution labels are applied.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
