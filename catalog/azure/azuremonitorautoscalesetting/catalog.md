# Azure Monitor Autoscale Setting

Deploys an Azure Monitor autoscale setting -- the rule book that automatically adds and removes instances of one scalable target (a VM Scale Set, an App Service plan, or any other capacity-bearing resource) based on metric rules and schedules. A setting holds up to 20 profiles, exactly one in effect at any moment -- a matching fixed-date profile beats a matching recurrence profile, which beats the default -- and each profile carries a capacity envelope plus up to 10 metric rules; predictive autoscale (VM Scale Sets only) and email/webhook notifications round out the surface. Azure allows one setting per target, and the setting itself is free: you pay only for the instances it creates.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Autoscale setting** -- the scaling rule book bound to one target resource: capacity envelopes, metric rules, recurrence and fixed-date schedules, predictive autoscale, and scale-event notifications
- **Azure Tags** -- Planton-derived metadata tags merged with the manifest's `tags` (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A resource group** -- the setting lives in a referenced resource group.
- **A scalable target** -- the resource the setting controls, referenced by its ARM ID output (e.g. an AzureVirtualMachineScaleSet's `scale_set_id` or an AzureServicePlan's `service_plan_id`).

### Azure Subscription

- **One autoscale setting per target** -- Azure rejects a second setting on the same resource at deploy time.
- **The setting must live in the target's region** -- autoscale is evaluated regionally.
- **The target's tier must support autoscale** -- e.g. App Service plans need Standard tier or above (Basic plans are rejected at deploy time).
- **The maximum instance count is also bounded by the subscription's core quota** -- the 0-1000 spec bounds mirror the provider's static validation only.

## Deploy

### Console

Open the deployment store, find **Azure Monitor Autoscale Setting**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields: the target resource, the profiles with their envelopes and rules, and notifications. Start from the **CPU-Based Scale Set Autoscale** or **Business-Hours Schedule** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorAutoscaleSetting
metadata:
  name: web-pool-autoscale
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: prod-web
  name: web-pool-autoscale
  region: eastus
  targetResourceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-web/providers/Microsoft.Compute/virtualMachineScaleSets/web-pool
  profiles:
    - name: default
      capacity:
        minimum: 2
        maximum: 10
        default: 3
      rules:
        - metricTrigger:
            metricName: Percentage CPU
            metricResourceId:
              value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-web/providers/Microsoft.Compute/virtualMachineScaleSets/web-pool
            timeGrain: PT1M
            statistic: Average
            timeWindow: PT10M
            timeAggregation: Average
            operator: GreaterThan
            threshold: 75
          scaleAction:
            direction: Increase
            type: ChangeCount
            value: 1
            cooldown: PT5M
        - metricTrigger:
            metricName: Percentage CPU
            metricResourceId:
              value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-web/providers/Microsoft.Compute/virtualMachineScaleSets/web-pool
            timeGrain: PT1M
            statistic: Average
            timeWindow: PT20M
            timeAggregation: Average
            operator: LessThan
            threshold: 25
          scaleAction:
            direction: Decrease
            type: ChangeCount
            value: 1
            cooldown: PT15M
```

```shell
planton apply -f autoscale-setting.yaml
```

This creates the classic elastic pool: the scale set grows one instance at a time on sustained CPU above 75% and shrinks conservatively below 25%, inside a 2-10 instance envelope. An Infra Job tracks the provisioning in real time.

### InfraChart

When the setting's dependencies deploy in the same InfraChart, wire them with ValueFromRef:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-web
      fieldPath: status.outputs.resource_group_name
  name: web-pool-autoscale
  region: eastus
  targetResourceId:
    valueFrom:
      kind: AzureVirtualMachineScaleSet
      name: web-pool
      fieldPath: status.outputs.scale_set_id
  profiles:
    - name: default
      capacity:
        minimum: 2
        maximum: 10
        default: 3
```

The InfraPipeline resolves the dependency graph, deploys the resource group and the scale set first, then binds the autoscale setting to it -- the rules' `metricResourceId` fields reference the same `scale_set_id` output when the metric source is the target itself.

## Key Configuration

These are the most important decisions when configuring an autoscale setting. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scale out eagerly, scale in lazily** -- flapping (out, in, out again) burns money and warm-up time. The standard defense is asymmetry: a scale-out rule with a short window and cooldown, paired with a scale-in rule with a LOWER threshold, a longer window, and a longer cooldown. Azure adds its own guard: it projects what the metric would look like after a scale-in and skips the action if it would immediately re-trigger a scale-out -- so a scale-in that "mysteriously never happens" usually means the in/out thresholds sit too close together.

**`capacity.default` is your metrics-outage posture** -- it is what Azure applies when the metric is UNAVAILABLE, and only if the current count is below it (metric loss never scales you in). Set it to the count that keeps the service healthy under normal peak, not the minimum: during a metrics outage you want to be over-provisioned, not at the floor.

**A schedule marks a START, not a window** -- a recurrence profile takes effect at its day/hour/minute and STAYS in effect until another profile's schedule begins; there is no end time. A business-hours profile therefore needs a partner profile at 18:00 that returns capacity to the overnight shape -- a lone scheduled profile silently becomes the permanent shape after its first activation.

**Scheduled profiles choose between pinning and re-enveloping** -- exactly one profile governs at any moment, and THAT profile's envelope and rules apply. A scheduled profile with no rules pins the count to its capacity default (right for predictable loads); a scheduled profile WITH rules gives a different elasticity envelope per time window (right when nights are quiet but not dead).

**Treat `ExactCount` as a deliberate jump** -- `ChangeCount` and `PercentChangeCount` nudge; `ExactCount` teleports. A misconfigured ExactCount rule (or a fixed-date profile with an aggressive default) is the classic accidental fleet-doubling -- review any ExactCount value against the profile's maximum before shipping it.

**Predictive autoscale: forecast first, act later** -- predictive mode (VM Scale Sets only, CPU-based) has a safe on-ramp: run `ForecastOnly` for a week and compare the forecast against reality before switching to `Enabled`. Use `lookAheadTime` (PT1M-PT1H) to cover your instances' real warm-up -- capacity that arrives exactly on time arrived late. There is no "Disabled" value: omit the whole `predictive` block to disable.

**Notifications name real inboxes, never "the administrator"** -- Azure retired the classic subscription-administrator email flags in April 2024 and ARM rejects any request that sets them, which is why this kind carries no such fields. Put an on-call alias in `notification.email.customEmails` (an email block needs at least one address) or wire a webhook to chat and paging systems; Azure fires them on every scale action.

**`enabled: false` freezes without forgetting** -- it stops evaluation and pins the current instance count while keeping the entire rule book in place, the right lever during incident response or load tests. Hand-editing rules in the portal instead drifts the IaC state: Azure allows one setting per target, so this manifest is the target's one authoritative rule book -- route every change through it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureVirtualMachineScaleSet** (scale-set targets) | `targetResourceId`, `metricTrigger.metricResourceId` | `status.outputs.scale_set_id` |
| **AzureServicePlan** (App Service plan targets) | `targetResourceId` | `status.outputs.service_plan_id` |

`targetResourceId` and `metricResourceId` carry no default kind because many kinds can be the target -- reference the resource's `*_id` output explicitly with valueFrom (kind + fieldPath), or pass a literal ARM ID.

### What This Kind Provides

This kind has no consumable outputs. `status.outputs` records the setting's ARM resource ID (`autoscale_setting_id`) and its name (`autoscale_setting_name`, which echoes the manifest's `name`), but an autoscale setting is a leaf: it acts on its target, and no catalog kind references a setting downstream.

## Common Patterns

**CPU-driven elastic pool** -- one default profile with asymmetric CPU rules (out at 75% over 10 minutes, in at 25% over 20 minutes with a longer cooldown) inside a modest envelope; the shape for any stateless scale-set workload whose load tracks CPU. Start from the **CPU-Based Scale Set Autoscale** preset.

**Calendar-shaped capacity** -- three profiles for an App Service plan: a business-hours recurrence profile, a partner profile that returns to the off-hours shape (because a schedule marks a start, not a window), and a schedule-less default covering weekends. Start from the **Business-Hours Schedule** preset.

**Queue-driven workers** -- point `metricResourceId` at a resource OTHER than the target (for example a queue's depth metric) so a worker scale set grows with backlog rather than CPU; `divideByInstanceCount` gives per-instance semantics on aggregate metrics.

**Launch-day surge window** -- a fixed-date profile with a raised envelope covering a one-off event (launch, sale weekend, migration); after the end timestamp, the default or recurring profile takes over again automatically.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the group the setting lives in
- [**Azure Virtual Machine Scale Set**](/infra-catalog/azure-virtual-machine-scale-set) -- the most common target, referenced by its `scale_set_id` output; also the only target predictive autoscale supports
- [**Azure Service Plan**](/infra-catalog/azure-service-plan) -- App Service plan targets (Standard tier or above), referenced by `service_plan_id`
