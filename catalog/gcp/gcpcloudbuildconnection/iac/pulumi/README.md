# GcpCloudBuildConnection — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud Build repository connection from the Planton spec: `gcp.projects.Service` and one `gcp.cloudbuildv2.Connection`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `connection` |
| `module/locals.go` | IaC input |
| `module/connection.go` | API enablement, the connection, one builder per code host, the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **One host block** -- each builder returns nil when its block is unset; optional fields set only when non-empty.
- **Installation outputs** -- the first `InstallationStates` entry's stage and action URI, empty when absent.
- **Name** -- set explicitly from `connectionId` (default `metadata.name`), never auto-named.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
