# GcpTpuQueuedResource — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud TPU queued resource from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.tpu.V2QueuedResource`. pulumi-gcp serves Google's beta-only resource from its single provider.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `queuedResource` |
| `module/locals.go` | Stack input holder |
| `module/queued_resource.go` | Resolves the node parent, enables the API, maps the node specs, exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `queued_resource_id`, `zone`) |

## Send Posture (parity with Terraform)

- **`Name`** -- `spec.queued_resource_id`, defaulting to `metadata.name`.
- **`Parent`** -- `projects/{project}/locations/{zone}`, the project from the spec or `organizations.GetClientConfig`.
- **Node optionals** -- only when set; the subnetwork trimmed to its relative path.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
