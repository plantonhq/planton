# GcpDeployPolicy — Terraform Implementation

This directory contains the Terraform implementation for a Cloud Deploy deploy policy from the Planton spec: `google_project_service` for the Cloud Deploy API and one `google_clouddeploy_deploy_policy`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Project, policy ID default, attribution labels |
| `main.tf` | `google_project_service` and the policy with its rules and selectors |
| `outputs.tf` | `name`, `deploy_policy_id`, `uid` |

## Send Posture

- **Optional fields** -- sent only when set; `suspended` only when true; empty `actions`, `invokers`, `daysOfWeek`, and selector labels are not sent.
- **Clock and date parts** -- each sent only when non-zero (Google reads an unset part as zero); a weekly window's times are blocks only when declared.
- **Labels** -- attribution labels merged into the policy's `labels`, never into a selector's labels (they are match criteria).

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
