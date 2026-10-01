# GCP Deploy Policy

Puts change freezes into code. A deploy policy blocks rollouts to the pipelines and targets you choose during the windows you choose -- every weekend, Friday afternoons, a year-end freeze -- for people, for automations, or both. A blocked rollout fails with a clear policy violation, and the people you trust can still override it in an emergency.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- the Cloud Deploy API on the policy's project
- **Deploy policy** -- the policy with its rules and selectors

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with permission to manage Cloud Deploy deploy policies in the target project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP Deploy Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Weekend Freeze** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDeployPolicy
metadata:
  name: prod-weekend-freeze
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-delivery
  location: us-central1
  rules:
    - rolloutRestriction:
        id: weekend-freeze
        timeWindows:
          timeZone: America/New_York
          weeklyWindows:
            - daysOfWeek:
                - SATURDAY
                - SUNDAY
  selectors:
    - target:
        labels:
          env: prod
```

```shell
planton apply -f deploy-policy.yaml
```

This blocks every rollout action on targets labeled `env: prod` all weekend, New York time. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a delivery pipeline's `status.outputs.delivery_pipeline_id` from `selectors[].deliveryPipeline.id`, and a target's `status.outputs.target_id` from `selectors[].target.id`.

## Key Configuration

These are the most important decisions when configuring a deploy policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**What it governs** -- selectors pick pipelines and targets by ID, by `*` for all of them in the location, or by labels. Any selector matching is enough.

**What it blocks** -- each rule's actions (creating rollouts, approving, rolling back, and more) and invokers (people, automations, or both). Empty lists block everything.

**When** -- weekly windows (days and hours) and one-time windows (dated), all in one time zone.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpDeliveryPipeline** | `selectors[].deliveryPipeline.id` | `status.outputs.delivery_pipeline_id` |
| **GcpDeployTarget** | `selectors[].target.id` | `status.outputs.target_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The policy's full resource name | IAM, audits |
| `deploy_policy_id` | The policy's ID | The name an emergency override gives |
| `uid` | The policy's unique identifier | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Weekend freeze** -- no production rollouts on Saturdays and Sundays. Start from the **Weekend Freeze** preset.

**Holiday freeze** -- a dated freeze plus Friday afternoons, for people only. Start from the **Holiday Freeze** preset.

## Works With

- [**GCP Delivery Pipeline**](/cloud-catalog/gcp-delivery-pipeline) -- pipelines the policy governs
- [**GCP Deploy Target**](/cloud-catalog/gcp-deploy-target) -- targets the policy governs
