# GcpCloudRunWorkerPool — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud Run worker
pool from the Planton spec: one `gcp.cloudrunv2.WorkerPool` with the Cloud
Run Admin API enabled.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `worker_pool` |
| `module/locals.go` | Pool name (spec or metadata.name fallback), the platform labels |
| `module/worker_pool.go` | Enables `run.googleapis.com`; maps spec to `gcp.cloudrunv2.WorkerPool` (template, containers, volumes, scaling, splits); exports the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`DeletionProtection`** -- always sent (spec default `true`).
- **`Scaling`** -- each lever sent only when set.
- **`InstanceSplits`** -- omitted when the spec list is empty.
- **`Resources.Limits`** -- `cpu` and `memory` as the two map keys; the
  block is omitted when neither is set.
- **Probes** -- one handler arm; no TCP liveness; the SDK's worker-pool
  probe `HttpHeaders` is a single header, which is why the spec caps the
  list at one -- both engines accept exactly the same manifests.
- **`SandboxLauncher`** -- the SDK's `WorkerPoolTemplateContainerArgs` has
  no such field (provider 8.1 added it); held out of the spec until the
  SDK carries it (re-evaluated at pulumi-gcp v10 GA).
- **`DeletionPolicy`** -- sent only when set.
- **`name`** -- the resource ID (the full resource name), the shape the
  Terraform module's `id` exports; `worker_pool_name` is the bare name.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
