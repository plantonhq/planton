# GcpKmsKeyHandle — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud KMS Autokey key handle from the Planton spec: `gcp.projects.Service` for the Cloud KMS API and one `gcp.kms.KeyHandle`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `keyHandle` |
| `module/locals.go` | IaC input holder |
| `module/key_handle.go` | Project resolution, API enablement, the handle, the outputs |
| `module/outputs.go` | Output key constants (`name`, `kms_key`) |

## Send Posture (parity with Terraform)

- **`Project`** -- the spec's project, or the provider's project from `organizations.GetClientConfig`.
- **`Name`** -- `key_handle_name`, defaulting to `metadata.name` (always sent, never auto-named).
- **Destroy** -- the provider removes the handle from state only; Google keeps it.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
