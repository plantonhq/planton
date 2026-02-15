# GcpCloudTasksQueue Terraform Module

Terraform implementation for provisioning a GCP Cloud Tasks queue.

## Provider

Requires `hashicorp/google` provider `~> 6.0`.

## Usage

```hcl
module "cloud_tasks_queue" {
  source = "."

  spec = {
    project_id = { value = "my-gcp-project" }
    queue_name = "my-task-queue"
    location   = "us-central1"

    rate_limits = {
      max_dispatches_per_second  = 500
      max_concurrent_dispatches = 100
    }

    retry_config = {
      max_attempts       = 5
      min_backoff        = "1s"
      max_backoff        = "3600s"
      max_doublings      = 16
    }
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| spec | GcpCloudTasksQueue spec | object | yes |
| provider_config | GCP provider configuration | object | no |

## Outputs

| Name | Description |
|------|-------------|
| queue_id | Fully qualified queue ID |
| queue_name | Short queue name |
| state | Current queue state |

## Notes

- Cloud Tasks queues do NOT support GCP labels.
- The `desired_state` field is a virtual Terraform field that calls pause/resume APIs separately from the standard PATCH update.
- `max_burst_size` in rate_limits is computed by GCP and cannot be set directly.
