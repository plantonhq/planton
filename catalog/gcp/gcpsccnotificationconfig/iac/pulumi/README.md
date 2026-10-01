# GcpSccNotificationConfig — Pulumi Implementation

This directory contains the Pulumi implementation for a Security Command Center notification config from the Planton spec: one of `gcp.securitycenter.V2ProjectNotificationConfig`, `V2FolderNotificationConfig`, or `V2OrganizationNotificationConfig`, selected by the scope, plus `gcp.projects.Service` for a project config.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and the resource function |
| `module/locals.go` | Stack input holder |
| `module/notification_config.go` | The scope switch, project resolution and API enablement, the resource, the outputs |
| `module/outputs.go` | Output key constants (`name`, `service_account`) |

## Send Posture (parity with Terraform)

- **Scope** -- a switch creates exactly one resource; the project arm or an empty scope (the provider's project from `organizations.GetClientConfig`) is the project resource.
- **`Location`** -- `spec.location`, defaulting to `global`.
- **`StreamingConfig.Filter`** -- always sent.
- **`PubsubTopic`** -- when set (required by the folder and organization resources).
- **`Description`**, **`DeletionPolicy`** -- only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
