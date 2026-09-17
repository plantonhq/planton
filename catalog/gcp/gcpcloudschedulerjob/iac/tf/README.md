# GcpCloudSchedulerJob Terraform Module

Provisions a Google Cloud Scheduler job using the Terraform Google provider (`~> 8.3`).

## Resources Created

- `google_project_service.cloudscheduler_api` -- Enables the Cloud Scheduler API
- `google_cloud_scheduler_job.this` -- The Cloud Scheduler job

## Usage

```hcl
module "scheduler_job" {
  source = "./path/to/module"

  metadata = {
    name = "nightly-trigger"
  }

  spec = {
    project_id = "my-project"
    location   = "us-central1"
    schedule   = "0 9 * * 1-5"
    http_target = {
      uri         = "https://my-service.run.app/api/trigger"
      http_method = "POST"
      oidc_token = {
        service_account_email = "invoker@my-project.iam.gserviceaccount.com"
      }
    }
  }
}
```

## Inputs

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `metadata` | object | yes | Resource metadata (name, labels) |
| `spec` | object | yes | GcpCloudSchedulerJob specification |

## Outputs

| Output | Description |
|--------|-------------|
| `job_id` | Fully qualified job ID |
| `job_name` | Short job name |
| `state` | Job state (ENABLED, PAUSED, DISABLED, UPDATE_FAILED) |

## Notes

- Requires Google provider `~> 8.3`
- Cloud Scheduler jobs do not support GCP labels
- If `job_name` is empty, the job takes its name from `metadata.name`
- If `project_id` is empty, the job is created in the provider's default
  project (ambient project contract)
- Dynamic blocks handle the three mutually exclusive target types
