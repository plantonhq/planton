# GcpSccMuteConfig — Pulumi Implementation

This directory contains the Pulumi implementation for a Security Command Center mute config from the Planton spec: one of `gcp.securitycenter.V2ProjectMuteConfig`, `V2FolderMuteConfig`, or `V2OrganizationMuteConfig`, selected by the scope, plus `gcp.projects.Service` for a project config.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and the resource function |
| `module/locals.go` | Stack input holder |
| `module/mute_config.go` | The scope switch, project resolution and API enablement, the resource, the outputs |
| `module/outputs.go` | Output key constants (`name`) |

## Send Posture (parity with Terraform)

- **Scope** -- a switch creates exactly one resource; the project arm or an empty scope (the provider's project from `organizations.GetClientConfig`) is the project resource.
- **`Location`** -- `spec.location`, defaulting to `global`.
- **`Filter`**, **`Type`** -- sent as declared.
- **`Description`**, **`DeletionPolicy`** -- only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
