# GcpColabSchedule — Pulumi Implementation

This directory contains the Pulumi implementation for a Vertex AI schedule from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.colab.Schedule`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `schedule` |
| `module/locals.go` | Stack input holder |
| `module/schedule.go` | Enables the API; maps the schedule, the notebook or pipeline request, and resolves a pipeline network's project number; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `schedule_id`, `location`) |

## Send Posture (parity with Terraform)

- **`DisplayName`** -- defaulting to `metadata.name`; the notebook run's to the schedule's.
- **Run counts** -- decimal strings; optional counts only when non-zero.
- **Requests** -- exactly one; `Parent` left to the provider.
- **`WorkbenchRuntime`** -- the empty marker, only when the bool is true.
- **Pipeline `Network`** -- trimmed and rebuilt with the project number through `organizations.LookupProject` when needed.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
