# GcpCloudBuildRepository — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud Build repository link from the Planton spec: one `gcp.cloudbuildv2.Repository`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `repository` |
| `module/locals.go` | Stack input |
| `module/repository.go` | The link, the connection-name parser, the outputs |
| `module/outputs.go` | Output key constants (`name`, `repository_id`, `remote_uri`) |

## Send Posture (parity with Terraform)

- **Placement** -- `parseConnectionName` splits project and location from `parentConnection`.
- **Name** -- set explicitly from `repositoryId` (default `metadata.name`), never auto-named.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
