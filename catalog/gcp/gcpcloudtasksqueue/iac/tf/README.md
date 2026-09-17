# GcpCloudTasksQueue Terraform Module

Terraform implementation for provisioning a GCP Cloud Tasks queue.

## Provider

Requires `hashicorp/google` provider `~> 8.3`.

## Usage

```hcl
module "cloud_tasks_queue" {
  source = "."

  metadata = {
    name = "my-task-queue"
  }

  spec = {
    project_id = "my-gcp-project"
    queue_name = "my-task-queue"
    location   = "us-central1"

    rate_limits = {
      max_dispatches_per_second = 500
      max_concurrent_dispatches = 100
    }

    retry_config = {
      max_attempts  = 5
      min_backoff   = "1s"
      max_backoff   = "3600s"
      max_doublings = 16
    }
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata (name, labels) | object | yes |
| spec | GcpCloudTasksQueue spec | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| queue_id | Fully qualified queue ID |
| queue_name | Short queue name |
| max_burst_size | Effective max burst size computed by GCP from the dispatch rate |

## Notes

- Cloud Tasks queues do NOT support GCP labels.
- If `project_id` is empty, the queue is created in the provider's default
  project (ambient project contract).
- `max_burst_size` in rate_limits is computed by GCP and cannot be set
  directly; the effective value is exported as an output.
- A deleted queue's ID is reserved by the Cloud Tasks API for up to 7 days;
  it cannot be reused within that window.
- Pause/resume (the API's queue state) is a runtime operation, not part of
  this declarative surface — use `gcloud tasks queues pause|resume`.
