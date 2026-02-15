# GcpCloudTasksQueue Pulumi Module - Architecture Overview

## Module Structure

The Pulumi module follows the standard OpenMCF GCP component pattern:

- **main.go** -- Entry point. Receives `GcpCloudTasksQueueStackInput`, initializes locals, creates the GCP provider, and delegates to resource functions.
- **locals.go** -- Computes derived values. Notably, Cloud Tasks queues do NOT support GCP labels, so no label map is computed (unlike most GCP components).
- **cloud_tasks_queue.go** -- Creates the `cloudtasks.NewQueue` resource with conditional blocks for all optional configurations.
- **outputs.go** -- Defines output key constants.

## Resource Flow

```
StackInput
  ├── target (GcpCloudTasksQueue manifest)
  └── provider_config (GCP credentials)
        │
        ▼
  initializeLocals()
        │
        ▼
  pulumigoogleprovider.Get()
        │
        ▼
  cloudTasksQueue()
    ├── Basic: name, location, project
    ├── desired_state (optional)
    ├── http_target (optional)
    │   ├── http_method
    │   ├── header_overrides[]
    │   ├── oauth_token OR oidc_token
    │   └── uri_override (with path/query flattening)
    ├── rate_limits (optional)
    ├── retry_config (optional)
    └── stackdriver_logging_config (optional)
        │
        ▼
  Exports: queue_id, queue_name, state
```

## Key Implementation Details

### No GCP Labels

Unlike most GCP components, Cloud Tasks queues do not accept labels. The `locals.go` file does not compute a label map, and no labels are passed to the resource.

### HTTP Target Flattening

The proto spec flattens `uri_override.path` and `uri_override.query_params` for cleaner UX. The Pulumi module maps these back to the SDK's nested structure:

- `spec.http_target.uri_override.path` -> `QueueHttpTargetUriOverridePathOverrideArgs.Path`
- `spec.http_target.uri_override.query_params` -> `QueueHttpTargetUriOverrideQueryOverrideArgs.QueryParams`

### Outputs

- `queue_id` -- Exported from `createdQueue.ID()` (fully qualified Pulumi resource ID)
- `queue_name` -- Exported from `createdQueue.Name`
- `state` -- Exported from `createdQueue.State` (RUNNING/PAUSED/DISABLED)

The `state` output is only available in Pulumi, not in Terraform.
