# Cloudflare Workflow

Deploys a Cloudflare Workflow registration: the binding of a durable-execution class exported by a deployed Worker script to a named workflow in the account, with instance retention, step limits, and cron triggers. Instances of the workflow are created at runtime by the Workers platform, never by this resource. Cloudflare's create is a PUT on the workflow name — registering an existing name silently adopts and overwrites it rather than failing.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Workflow Registration** — one workflow on the account binding `workflowName` to the `className` exported by `scriptName`, carrying the retention, limits, and schedule settings

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with an API token holding Account → Workers Scripts → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **A deployed Worker script** — `scriptName` must name a script already deployed (a CloudflareWorker Infra Component) that exports the `WorkflowEntrypoint` subclass named in `className`. Deploy order is script first, workflow second.
- **Workers Paid** (only for scale) — the registration itself works on the free tier; production instance runs at scale ride Workers Paid limits.

## Deploy

### Console

Open the deployment store, find **Cloudflare Workflow**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the account, the name/class/script binding, and the retention, limits, and schedule settings. Start from the **Scheduled workflow** preset in the [Presets](#presets) tab to pre-populate a nightly batch shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWorkflow
metadata:
  name: order-fulfillment
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  workflowName: order-fulfillment
  className: OrderFulfillment
  scriptName:
    value: order-processor
```

```shell
planton apply -f workflow.yaml
```

This registers the `order-fulfillment` workflow against the `OrderFulfillment` class exported by the already-deployed `order-processor` script, keeping Cloudflare's default retention and limits. An Infra Job tracks the provisioning in real time.

### InfraChart

When the Worker is deployed in the same InfraPipeline, wire `scriptName` with ValueFromRef:

```yaml
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  workflowName: order-fulfillment
  className: OrderFulfillment
  scriptName:
    valueFrom:
      kind: CloudflareWorker
      name: order-processor
      fieldPath: status.outputs.script_name
```

The InfraPipeline resolves the dependency graph, deploys the Worker first, then registers the workflow against the resolved script name — the script-before-workflow ordering for free.

## Key Configuration

These are the most important decisions when configuring a workflow. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The name is an upsert key** — Cloudflare has no separate create endpoint; registration writes `PUT` to the workflow name. Registering a name that already exists silently adopts and overwrites the existing workflow's binding (class, script, retention, schedules), and neither apply ever errors. Two teams sharing a name will fight over one registration. Namespace workflow names deliberately, exactly as you would Worker script names. Changing `workflowName` (or `accountId`) replaces the workflow; `className` and `scriptName` update in place.

**The registration is not the program** — a typo in `className` may register fine and fail only when the first instance runs, because Cloudflare validates the binding lazily. Verify the class name against your script's exports, not the apply result.

**Retention takes two forms** — `defaultRetention.errorRetention` and `successRetention` accept integer milliseconds (`"86400000"`) or duration expressions (`"1 day"`). Cloudflare normalizes internally and equivalent values do not diff, but mixed forms across manifests make review harder. Prefer the duration-expression form for anything a human reads. A longer error retention than success retention buys a debugging window without paying for successful-run state.

**Schedules create instances** — each `schedules[].cron` entry (five-field cron, evaluated in UTC) starts a new instance on its cadence. `limits.steps` caps what any one instance may execute (at least 1).

**Instances outlive your apply** — destroying the registration does not await or terminate running instances; they belong to the platform. To drain before decommissioning: remove `schedules`, stop creating instances from your Worker, let instances finish, then destroy. The API keeps answering GET for deleted workflows with an `is_deleted` marker instead of a 404 — the record is a tombstone, not a survivor.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareWorker** | `scriptName` | `status.outputs.script_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `workflow_name` | The workflow's identity within the account | Worker `workflows` bindings that create instances of this workflow |
| `version_id` | The workflow version the registration produced | Version pinning and audit of what a deploy registered |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Scheduled batch workflow** — a nightly multi-step job (reconciliation, report generation, cleanup sweeps) triggered by cron, with a week of error retention for debugging and a day for successes. Start from the **Scheduled workflow** preset.

**Event-driven workflow** — no `schedules`; instances are created by the Worker itself (often fed by a CloudflareQueue) through a `workflows` binding that references this registration by name. The registration carries only retention and limits.

## Works With

- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) — the script that exports the workflow class; its `workflows` bindings reference this registration by name
- [**Cloudflare Queue**](/infra-catalog/cloudflare-queue) — queues feed Workers that create workflow instances; the sibling asynchronous-work building block
