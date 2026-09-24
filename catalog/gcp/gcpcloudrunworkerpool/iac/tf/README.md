# GcpCloudRunWorkerPool — Terraform Implementation

This directory contains the Terraform implementation for a Cloud Run worker
pool from the Planton spec: one `google_cloud_run_v2_worker_pool` with the
Cloud Run Admin API enabled.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, pool name (spec or metadata.name fallback), null-when-empty levers, the platform labels, the VPC access guards |
| `main.tf` | `google_project_service`, `google_cloud_run_v2_worker_pool` |
| `secrets.tf` | Per `env[].secret_value`: the Secret Manager API, one secret replicated in the pool's region, its version, and the runtime identity's `secretAccessor` grant on that secret alone; the Compute Engine default identity lookup when `service_account` is unset |
| `outputs.tf` | `name`, `worker_pool_name`, `uid`, `location`, `project_id`, `latest_created_revision`, `latest_ready_revision`, `observed_generation`, `etag` |

## Send Posture

- **`deletion_protection`** -- always sent (spec default `true`).
- **`scaling`** -- each lever sent only when set, so Google's MANUAL default
  and default count stay in charge otherwise.
- **`instance_splits`** -- omitted when the spec list is empty, so Google
  routes every instance to the latest ready revision without a diff-prone
  split.
- **`template.containers[].resources.limits`** -- the spec's `cpu` and
  `memory` land as the two keys of the API's limits map; null (not `{}`)
  when neither is set. A worker pool has no `cpu_idle` or
  `startup_cpu_boost`: CPU is always allocated.
- **Probes** -- exactly one handler arm arrives (proto-oneof); the
  liveness probe has no TCP arm (Cloud Run rejects it); at most one
  `http_headers` entry (spec-capped, because the pinned Pulumi SDK models
  a worker-pool probe's headers as one header -- PARITY).
- **`sandbox_launcher`** -- not spec surface: the pinned Pulumi SDK's
  worker-pool container type lacks it (re-evaluated at pulumi-gcp v10 GA).
- **`deletion_policy`** -- DELETE (default), PREVENT, or ABANDON; sent only
  when set.
- **`env[]`** -- exactly one of `value`, `value_source` from
  `value_from_secret`, or `value_source` pointing at the module's own secret
  for `secret_value`, pinned to the stored version. Secret ids are
  `runpool_<region>_<pool>_<container>_<variable>` (an unnamed container is
  `c<index>`, a `.` in a name becomes `-`), the same law the Pulumi module's
  shared helper applies, so both engines name the same secrets. The pool
  waits for every grant, because Cloud Run checks the runtime identity's
  access when it creates a revision.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
