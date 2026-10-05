# GcpCloudBuildWorkerPool — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud Build private worker pool from the Planton spec: `gcp.projects.Service` and one `gcp.cloudbuild.WorkerPool`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `workerPool` |
| `module/locals.go` | IaC input |
| `module/worker_pool.go` | API enablement, the pool, the peered-network resolver, the outputs |
| `module/outputs.go` | Output key constants (`name`, `worker_pool_id`, `state`, `uid`) |

## Send Posture (parity with Terraform)

- **Optional fields** -- set only when declared; the worker booleans whenever present, including `false`.
- **Peered network** -- `resolvePeeredNetwork` strips the self-link prefix and resolves a project ID to its number with one `organizations.LookupProject`.
- **Name** -- set explicitly from `workerPoolId` (default `metadata.name`), never auto-named.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
