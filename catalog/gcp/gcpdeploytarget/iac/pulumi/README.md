# GcpDeployTarget — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud Deploy target from the Planton spec: `gcp.projects.Service` and one `gcp.clouddeploy.Target`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `target` |
| `module/locals.go` | IaC input, attribution labels, the label merge |
| `module/target.go` | API enablement, the target, one args helper per nested block, the outputs |
| `module/outputs.go` | Output key constants (`name`, `target_id`, `uid`) |

## Send Posture (parity with Terraform)

- **Optional fields** -- `optionalString` and `optionalTrue` send a value only when it is set; a nested-block helper returns `nil` when its block is absent.
- **Target type** -- only the declared type's block is sent.
- **Name** -- set explicitly from `targetId` (default `metadata.name`), never auto-named.
- **Labels** -- attribution labels merged on top of the spec's labels.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
