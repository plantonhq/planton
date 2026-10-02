# GcpCloudBuildRepository — Terraform Implementation

This directory contains the Terraform implementation for a Cloud Build repository link from the Planton spec: one `google_cloudbuildv2_repository`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Project and location parsed from the connection's full name, the repository ID default |
| `main.tf` | The repository link |
| `outputs.tf` | `name`, `repository_id`, `remote_uri` |

## Send Posture

- **Placement** -- `project` and `location` split from `parentConnection` -- PARITY with the Pulumi module's `parseConnectionName`.
- **No API enablement** -- the parent connection's block enabled Cloud Build.
- **No labels** -- the link has annotations only.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
