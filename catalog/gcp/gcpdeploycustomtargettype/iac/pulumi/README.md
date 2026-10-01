# GcpDeployCustomTargetType — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud Deploy custom target type from the Planton spec: `gcp.projects.Service` and one `gcp.clouddeploy.CustomTargetType`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `customTargetType` |
| `module/locals.go` | Stack input, attribution labels, the label merge |
| `module/custom_target_type.go` | API enablement, the type, the custom-actions and tasks builders, the send-only-when-set helpers, the outputs |
| `module/outputs.go` | Output key constants (`name`, `custom_target_type_id`, `uid`) |

## Send Posture (parity with Terraform)

- **One definition** -- each builder returns nil when its block is unset; a Skaffold module carries only the source it declares; a task's container only when declared.
- **Container command** -- the spec's `command` is the SDK's `Commands` (Pulumi pluralizes the list); the provider receives `command`.
- **Labels** -- attribution labels on `Labels`.
- **Name** -- set explicitly from `customTargetTypeId` (default `metadata.name`), never auto-named.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
