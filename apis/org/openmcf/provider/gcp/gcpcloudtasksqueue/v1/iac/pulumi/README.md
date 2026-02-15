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

- Queue-level HTTP target with OIDC/OAuth authentication
- URI overrides (scheme, host, port, path, query params)
- Header overrides for all tasks
- Configurable rate limits and retry behavior
- Stackdriver logging with sampling ratio
- Desired state control (RUNNING/PAUSED)
- Queue state exported as output

## Notes

- Cloud Tasks queues do NOT support GCP labels. No labels are computed or applied.
- The `state` output is available via Pulumi but not via Terraform.
- `max_burst_size` is computed by GCP and exported via Pulumi's rate_limits output.
- Flattened URI path/query overrides are mapped back to the SDK's nested structure.

## Debug

```bash
cd ~/scm/github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpcloudtasksqueue/v1/iac/pulumi
make build
PULUMI_CONFIG_PASSPHRASE="" pulumi login --local
pulumi stack init dev
pulumi config set --path 'stack_input' --secret < ../../hack/manifest.yaml
make preview
```
