# CloudflareWorkflow

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareWorkflowSpec registers a Workflow: a durable, multi-step
execution program served by a class exported from a deployed Worker
script. The Workflow object binds a name to that class and carries the
engine settings (instance retention, step limits, cron triggers);
instances of the workflow are then created at runtime by the Worker
platform, never by this resource.

Two provider behaviors worth knowing before authoring:
  - Create IS a PUT: registering a workflow name that already exists
    silently adopts and overwrites it (the name-as-upsert class). Choose
    names deliberately.
  - The API keeps answering for deleted workflows with an is_deleted
    marker instead of a 404 -- deletion is real, but tooling that probes
    for existence must read the marker.

## Example

```yaml
# Complete example manifest for CloudflareWorkflow. Registers a durable
# workflow served by the OrderFulfillment class exported from a deployed
# Worker script, with explicit retention, a step cap, and a nightly cron
# trigger. The script must be deployed before the workflow registers.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWorkflow
metadata:
  name: order-fulfillment
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  workflow_name: order-fulfillment
  class_name: OrderFulfillment
  script_name:
    value: order-processor
  default_retention:
    error_retention: "7 days"
    success_retention: "86400000"
  limits:
    steps: 512
  schedules:
    - cron: "0 3 * * *"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.workflowName` | `string` | yes |  |  |
| `spec.className` | `string` | yes |  |  |
| `spec.scriptName` | `string \| valueFrom` | yes |  | CloudflareWorker (`status.outputs.script_name`) |
| `spec.defaultRetention` | `CloudflareWorkflowRetention` |  |  |  |
| `spec.defaultRetention.errorRetention` | `string` |  |  |  |
| `spec.defaultRetention.successRetention` | `string` |  |  |  |
| `spec.limits` | `CloudflareWorkflowLimits` |  |  |  |
| `spec.limits.steps` | `int64` |  |  |  |
| `spec.schedules` | `[]CloudflareWorkflowSchedule` |  |  |  |
| `spec.schedules[].cron` | `string` | yes |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the workflow lives in.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.workflowName

`string` · required

The workflow's name -- its identity within the account. Changing it
replaces the workflow (the provider forces a new resource). Because the
API treats create as an upsert, a name collision overwrites the existing
workflow rather than failing.

- rule: {"string":{"minLen":"1"}}

### spec.className

`string` · required

The name of the class exported by the Worker script that implements
this workflow (the `WorkflowEntrypoint` subclass).

- rule: {"required":true}

### spec.scriptName

`string | valueFrom` · required

The Worker script that exports the workflow class: a literal script
name, or a reference to a CloudflareWorker resource's script_name
output. The script must be deployed before the workflow registers.

- references: CloudflareWorker (`status.outputs.script_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareWorker, name: <that resource's name>, fieldPath: status.outputs.script_name}} -- a bare string does not parse

### spec.defaultRetention

`CloudflareWorkflowRetention`

How long finished workflow instances are retained before the engine
discards their state. Omit to keep Cloudflare's defaults.

### spec.defaultRetention.errorRetention

`string`

Retention for instances that finished in error, as milliseconds
("5000") or a duration expression ("5 minutes").

### spec.defaultRetention.successRetention

`string`

Retention for instances that finished successfully, as milliseconds
("5000") or a duration expression ("5 minutes").

### spec.limits

`CloudflareWorkflowLimits`

Engine limits applied to every instance of this workflow. Omit to keep
Cloudflare's defaults.

### spec.limits.steps

`int64` · optional (explicit presence)

The maximum number of steps an instance may execute. Cloudflare
requires at least 1.

- rule: {"int64":{"gte":"1"}}

### spec.schedules

`[]CloudflareWorkflowSchedule`

Cron triggers that start new workflow instances on a schedule. Each
entry is one cron expression evaluated in UTC.

### spec.schedules[].cron

`string` · required

The cron expression (five-field, evaluated in UTC), e.g. "0 3 * * *"
for 03:00 daily.

- rule: {"required":true}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareWorkflow, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.workflow_name` | `string` | The workflow's name -- its identity within the account, and what Worker workflow bindings reference. |
| `status.outputs.version_id` | `string` | The ID of the workflow version the registration produced. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.scriptName` | CloudflareWorker | `status.outputs.script_name` |

## See Also

- [Overview](../README.md)
