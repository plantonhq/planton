# GcpDeployCustomTargetType — Terraform Implementation

This directory contains the Terraform implementation for a Cloud Deploy custom target type from the Planton spec: `google_project_service` for the Cloud Deploy API and one `google_clouddeploy_custom_target_type`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Project, type ID default, attribution labels |
| `main.tf` | `google_project_service` and the custom target type with its custom actions or tasks |
| `outputs.tf` | `name`, `custom_target_type_id`, `uid` |

## Send Posture

- **Optional fields** -- sent only when set; empty `configs`, `command`, `args`, and `env` are not sent.
- **One definition** -- `custom_actions` or `tasks` blocks only when declared; each Skaffold module renders exactly the one source block it declares; a task's `container` only when declared.
- **Labels** -- attribution labels merged into `labels`.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
