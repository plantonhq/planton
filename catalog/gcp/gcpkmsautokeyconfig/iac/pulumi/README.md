# GcpKmsAutokeyConfig — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud KMS Autokey configuration from the Planton spec: `gcp.projects.Service` for the Cloud KMS API, and one `gcp.kms.AutokeyConfig` (folder) or `gcp.kms.ProjectAutokeyConfig` (project).

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `autokeyConfig` |
| `module/locals.go` | Stack input holder |
| `module/autokey_config.go` | Scope selection, API enablement, the configuration, the outputs |
| `module/outputs.go` | Output key constants (`name`, `parent`) |

## Send Posture (parity with Terraform)

- **Scope** -- a folder arm creates `AutokeyConfig` with the bare folder ID; otherwise `ProjectAutokeyConfig`, on the project arm's project or the provider's project from `organizations.GetClientConfig`.
- **`KeyProject`** -- `projects/{id}` when set.
- **`KeyProjectResolutionMode`**, **`DeletionPolicy`** -- only when set.
- **APIs** -- `cloudkms.googleapis.com` on the configuration's project and the key project; `DisableOnDestroy` false.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
