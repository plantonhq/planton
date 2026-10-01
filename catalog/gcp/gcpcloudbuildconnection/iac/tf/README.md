# GcpCloudBuildConnection — Terraform Implementation

This directory contains the Terraform implementation for a Cloud Build repository connection from the Planton spec: `google_project_service` for the Cloud Build API and one `google_cloudbuildv2_connection`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Project and connection ID defaults |
| `main.tf` | `google_project_service` and the connection with its five code-host blocks |
| `outputs.tf` | `name`, `connection_id`, `installation_stage`, `installation_action_uri` |

## Send Posture

- **One host block** -- each rendered only when declared; optional host fields sent only when set.
- **Secrets** -- version names only, passed through as resolved by the tfvars converter.
- **Installation outputs** -- read from `installation_state[0]` with `try`, empty when Google reports none -- PARITY with the Pulumi module.
- **No labels** -- the connection has annotations only.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
