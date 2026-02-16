# GcpCloudSchedulerJob Terraform Module

Provisions a Google Cloud Scheduler job using the Terraform Google provider (`~> 6.0`).

## Resources Created

- `google_cloud_scheduler_job.this` -- The Cloud Scheduler job

## Usage

```hcl
module "scheduler_job" {
  source = "./path/to/module"

  spec = {
    project_id = { value = "my-project" }
    location   = "us-central1"
    schedule   = "0 9 * * 1-5"
    http_target = {
      uri         = "https://my-service.run.app/api/trigger"
      http_method = "POST"
      oidc_token = {
        service_account_email = { value = "invoker@my-project.iam.gserviceaccount.com" }
      }
    }
  }
}
```

## Inputs

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `spec` | object | yes | GcpCloudSchedulerJob specification |
| `provider_config` | object | no | GCP provider credentials |

## Outputs

| Output | Description |
|--------|-------------|
| `job_id` | Fully qualified job ID |
| `job_name` | Short job name |
| `state` | Job state (ENABLED, PAUSED, etc.) |

## Notes

- Requires Google provider `~> 6.0`
- Cloud Scheduler jobs do not support GCP labels
- The `state` output is available in Terraform (unlike Cloud Tasks where it was Pulumi-only)
- Dynamic blocks handle the three mutually exclusive target types
