# Cloudflare Workflow

## Overview

`CloudflareWorkflow` registers a Workflow: a durable, multi-step execution program served by a class exported from a deployed Worker script. The registration binds a workflow name to that class and carries the engine settings -- instance retention, step limits, and cron triggers. Workflow INSTANCES are then created at runtime by the Workers platform (or your code), never by this resource.

Two provider behaviors are worth knowing before authoring: create IS a PUT (registering an existing name silently adopts and overwrites it, never a conflict), and the API keeps answering GET for deleted workflows with an `is_deleted` marker instead of a 404 -- deletion is real, but existence probes must read the marker.

## Key Features

- **Durable execution** -- multi-step programs that survive restarts, with engine-managed retries and state
- **Bound to your Worker** -- `script_name` references a `CloudflareWorker` (or a literal script name); the class must be exported by that script
- **Retention control** -- how long finished instances stay queryable, as milliseconds ("5000") or duration expressions ("5 minutes")
- **Cron triggers** -- start instances on a schedule (UTC)

## Use Cases

**Ideal for:**

- Order pipelines, provisioning flows, and other multi-step jobs that must survive failures mid-run
- Scheduled batch work (a nightly reconcile) via cron triggers
- Long-running orchestrations Workers alone cannot hold state for

**Not ideal for:**

- The Worker script itself -- that is `CloudflareWorker` (this resource only registers the workflow on an already-deployed script)
- Simple request/response logic with no steps or durability needs

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `workflow_name` | string | Yes | The workflow's identity. Changing it replaces the workflow; a collision with an existing name silently adopts it (create is a PUT). |
| `class_name` | string | Yes | The `WorkflowEntrypoint` subclass the script exports. |
| `script_name` | StringValueOrRef | Yes | The deployed Worker script exporting the class (references a `CloudflareWorker`'s `script_name` output). |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `default_retention` | object | `error_retention` / `success_retention` -- how long finished instances stay queryable, as milliseconds ("5000") or a duration expression ("5 minutes"). |
| `limits` | object | `steps` -- the maximum steps an instance may execute (>= 1). |
| `schedules[]` | list | Cron triggers (`cron`, five-field, UTC). |

### Outputs

| Field | Description |
|-------|-------------|
| `workflow_name` | The workflow's identity -- what Worker workflow bindings reference |
| `version_id` | The workflow version the registration produced |

## Example Manifest

```yaml
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
  limits:
    steps: 512
  schedules:
    - cron: "0 3 * * *"
```

## Destroy Semantics

Destroy is a real delete, but the API keeps answering GET for the deleted workflow with a non-zero `is_deleted` marker instead of a 404. Running instances belong to the platform, not this resource -- deletion does not wait for them.

## Plan Notes

Registration itself works on the free Workers tier; production instance runs at scale ride Workers Paid limits. Verify the entitlement against your account's plan.

## Related Resources

- **CloudflareWorker** -- the script that exports the workflow class (and whose workflow bindings reference this registration by name)
- **CloudflareQueue** -- the other Workers-platform building block for asynchronous work

## Further Reading

For operational judgment -- the name-as-upsert trap, retention forms, why instances outlive the registration -- see GUIDE.md.

## References

- [Cloudflare Workflows](https://developers.cloudflare.com/workflows/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
