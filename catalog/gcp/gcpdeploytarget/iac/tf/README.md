# GcpDeployTarget — Terraform Implementation

This directory contains the Terraform implementation for a Cloud Deploy target from the Planton spec: `google_project_service` for the Cloud Deploy API and one `google_clouddeploy_target`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Project, target ID default, attribution labels |
| `main.tf` | `google_project_service`, the target with its type, associated entities, and execution configurations |
| `outputs.tf` | `name`, `target_id`, `uid` |

## Send Posture

- **Optional fields** -- sent only when set, so Cloud Deploy's defaults apply otherwise; booleans are sent only when true (false is the provider's default).
- **Target type** -- one dynamic block per type, rendered only for the type the spec declares.
- **Execution configurations** -- optional and computed: none declared sends none, and Google's defaults read back without a diff.
- **Labels** -- attribution labels merged on top of the spec's labels -- PARITY with the Pulumi module's `mergeLabels`.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
