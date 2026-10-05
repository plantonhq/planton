# GcpCloudSchedulerJob Pulumi Module

Provisions a Google Cloud Scheduler job using the Pulumi GCP provider.

## Architecture

The module enables the Cloud Scheduler API (`projects.Service`), then creates a
`cloudscheduler.Job` resource with conditional configuration for the selected
target type (HTTP, Pub/Sub, or App Engine) and optional retry configuration.

## Files

| File | Purpose |
|------|---------|
| `module/main.go` | Entry point - initializes locals, provider, and calls resource function |
| `module/locals.go` | Locals struct and initialization - extracts target and provider config |
| `module/cloud_scheduler_job.go` | Core resource creation with all target types and retry config |
| `module/outputs.go` | Output constant definitions (job_id, job_name, state) |
| `main.go` | Pulumi entrypoint - loads IaC input and delegates to module |

## Usage

```bash
make build   # Compile the Pulumi binary
make preview # Preview changes
make up      # Apply changes
make destroy # Tear down resources
```

## Notes

- Cloud Scheduler jobs do **not** support GCP labels
- The `state` output is computed by GCP (ENABLED, PAUSED, DISABLED, UPDATE_FAILED)
- Job name defaults to `metadata.name` if `job_name` is not specified
- If `project_id` is empty, the job lands in the provider's default project
- The `paused` field creates the job in PAUSED state; omitting it creates ENABLED
- Body fields are passed through as-is (expected to be base64-encoded by the user)
- The `deletion_policy` spec field (DELETE / PREVENT / ABANDON, provider default
  DELETE) is sent only when set — wired identically to the Terraform module
