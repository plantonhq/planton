# GcpColabSchedule — Terraform Implementation

This directory contains the Terraform implementation for a Vertex AI schedule from the Planton spec: one `google_project_service` (API enablement), one `google_colab_schedule`, and a guarded `data.google_project` read for a pipeline network whose path carries a project ID.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, display name defaults, run counts as strings, subnetwork trim, the pipeline network's project-number path |
| `main.tf` | `google_project_service`, `data.google_project` (count-gated), `google_colab_schedule` |
| `outputs.tf` | `name`, `schedule_id`, `location` |

## Send Posture

- **`display_name`** -- defaulting to `metadata.name`; the notebook run's display name defaults to the schedule's -- PARITY with the Pulumi module.
- **Run counts** -- the spec's `int64`s as decimal strings; `max_run_count` and `max_concurrent_active_run_count` only when non-zero.
- **Requests** -- exactly one of the two request blocks; each request's `parent` is left to the provider (it fills `projects/{project}/locations/{location}`).
- **`workbench_runtime`** -- the empty marker block, rendered only when the spec's bool is true.
- **Pipeline `network`** -- a self-link is trimmed and, when its project segment is not numeric, rebuilt with the project number from the guarded read.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
