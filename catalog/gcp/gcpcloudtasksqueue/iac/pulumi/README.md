# GcpCloudTasksQueue Pulumi Module

Pulumi Go implementation for provisioning a GCP Cloud Tasks queue.

## Architecture

```
module/
  main.go                 # Entry point (Resources function)
  locals.go               # Variable transformations
  cloud_tasks_queue.go    # Queue resource creation
  outputs.go              # Output constant definitions
```

## SDK

Uses `pulumi-gcp/sdk/v9/go/gcp/cloudtasks` for the `cloudtasks.NewQueue` resource.

## Features

- Cloud Tasks API enablement (fresh projects work first try)
- Queue-level HTTP target with OIDC/OAuth authentication
- URI overrides (scheme, host, port, path, query params)
- Header overrides for all tasks
- App Engine routing override (service/version/instance pinning)
- Configurable rate limits and retry behavior
- Stackdriver logging with sampling ratio

## Notes

- Cloud Tasks queues do NOT support GCP labels. No labels are computed or applied.
- If `project_id` is empty, the queue lands in the provider's default project.
- `max_burst_size` is computed by GCP from the dispatch rate and exported as an output.
- Flattened URI path/query overrides are mapped back to the SDK's nested structure.
- The `deletion_policy` spec field (DELETE / PREVENT / ABANDON, provider default
  DELETE) is sent only when set; `desired_state` (RUNNING / PAUSED) is sent
  explicitly on every apply so pause/resume reconciles declaratively — both
  mirror the Terraform module.

## Debug

```bash
cd ~/scm/github.com/plantonhq/planton/catalog/gcp/gcpcloudtasksqueue/iac/pulumi

# Preview against the kind's validated example manifest
planton pulumi preview \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```
