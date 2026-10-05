# Azure Monitor Metric Alert

Deploys an Azure Monitor metric alert rule -- the watchdog on platform metrics. It evaluates a metric (CPU, latency, queue depth, transactions -- anything a resource emits to Azure Monitor Metrics) against a condition on a rolling window and fires action groups when the condition holds. Three condition families exist: static thresholds (the classic "metric crosses a value"), dynamic machine-learning thresholds (Azure learns the metric's normal band), and web-test availability (fires when an Application Insights availability test fails from N locations -- how a failed probe becomes a page).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Metric Alert Rule** -- a `Microsoft.Insights/metricAlerts` resource (GLOBAL -- no region) carrying the scopes, exactly one condition family, the evaluation cadence, severity, and the action-group wiring
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically, merged with any user tags (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the alert RULE will be created (independent of where the watched resources live). Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **Something to watch** -- the scoped resources must exist and emit the metric; for the web-test family, the Application Insights standard web test and its component.
- **An action group** to notify (optional but recommended) -- a rule with no actions records state but notifies nobody.

## Deploy

### Console

Open the deployment store, find **Azure Monitor Metric Alert**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Static Threshold Alert** preset in the [Presets](#presets) tab to pre-populate the classic threshold rule.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorMetricAlert
metadata:
  name: checkout-latency-alert
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "observability-rg"
  alertName: checkout-api-high-latency
  scopes:
    - value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/web-rg/providers/Microsoft.Web/sites/checkout-api"
  description: p95 latency above 500ms -- runbook wiki/checkout-latency
  severity: 1
  staticCriteria:
    - metricNamespace: Microsoft.Web/sites
      metricName: HttpResponseTime
      aggregation: AVERAGE
      operator: GREATER_THAN
      threshold: 0.5
  actions:
    - actionGroupId:
        valueFrom:
          kind: AzureMonitorActionGroup
          name: platform-oncall
          fieldPath: status.outputs.action_group_id
```

```shell
planton apply -f metric-alert.yaml
```

This creates a rule on the platform defaults: evaluated every minute over the last five minutes, stateful (one firing per incident, self-resolving). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the rule to resources deployed in the same InfraPipeline -- the web-test availability family is the canonical example:

```yaml
spec:
  scopes:
    - valueFrom:
        kind: AzureApplicationInsightsStandardWebTest
        name: checkout-availability
        fieldPath: status.outputs.web_test_id
  webTestAvailabilityCriteria:
    webTestId:
      valueFrom:
        kind: AzureApplicationInsightsStandardWebTest
        name: checkout-availability
        fieldPath: status.outputs.web_test_id
    componentId:
      valueFrom:
        kind: AzureApplicationInsights
        name: checkout-api-insights
        fieldPath: status.outputs.application_insights_id
    failedLocationCount: 3
```

The InfraPipeline resolves the dependency graph, deploys the web test first, then provisions the alert with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring a metric alert. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Exactly one condition family** -- `staticCriteria` (one or more thresholds -- ALL must breach), `dynamicCriteria` (one ML criterion with sensitivity and the failing-periods flap damper), or `webTestAvailabilityCriteria` (N simultaneous location failures). Static operators compare in one direction; both-directions deviation (GREATER_OR_LESS_THAN) belongs to the dynamic family.

**Scopes** -- ARM IDs of what the rule watches: a resource, a resource group, or a subscription. Multiple scopes (or a group/subscription scope) require `targetResourceType` + `targetResourceLocation` so Azure can resolve the metric definition -- multi-resource rules evaluate each resource independently.

**Cadence** -- `frequency` (PT1M...PT1H; default every minute) and `windowSize` (PT1M...P1D; default five minutes). The window must be at least the frequency. `severity` is 0 (critical) through 4 (verbose), defaulting to 3 -- zero is a real choice, raise deliberately for paging conditions.

**Stateful behavior** -- `autoMitigate` (Azure's default true) fires once per incident and self-resolves; false fires on every breaching evaluation. `enabled: false` pauses evaluation for maintenance without deleting the rule.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **Any metric-emitting kind** | `scopes[]` | the watched resource's `*_id` output (explicit valueFrom) |
| **AzureApplicationInsightsStandardWebTest** | `webTestAvailabilityCriteria.webTestId` | `status.outputs.web_test_id` |
| **AzureApplicationInsights** | `webTestAvailabilityCriteria.componentId` | `status.outputs.application_insights_id` |
| **AzureMonitorActionGroup** | `actions[].actionGroupId` | `status.outputs.action_group_id` |

### What This Kind Provides

The rule is a leaf in the dependency graph: `status.outputs` carries only its own identifiers (`metric_alert_id`, `metric_alert_name`) for filtering alert history and CLI reference -- no downstream Infra Component consumes them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Static threshold** -- the classic paging rule: an average crosses a number you know means trouble. Start from the **Static Threshold Alert** preset.

**Dynamic anomaly** -- Azure learns the metric's seasonal normal and alerts on deviation -- for request volume, queue depth, and anything where a fixed number is wrong half the day. Start from the **Dynamic Anomaly Alert** preset.

**Web-test availability** -- the outside-in pager: fires when the availability probe fails from three or more locations. Start from the **Web-Test Availability Alert** preset.

## Works With

- [**Azure Monitor Action Group**](/infra-catalog/azure-monitor-action-group) -- the notification hub the rule fires into
- [**Azure Application Insights Standard Web Test**](/infra-catalog/azure-application-insights-standard-web-test) -- the availability probe the web-test family watches
- [**Azure Application Insights**](/infra-catalog/azure-application-insights) -- the component behind the web test
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the rule is created
