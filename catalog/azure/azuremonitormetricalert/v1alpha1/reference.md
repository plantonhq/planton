# AzureMonitorMetricAlert

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureMonitorMetricAlertSpec** defines the configuration for creating an
Azure Monitor metric alert rule.

A metric alert evaluates a platform metric (CPU, latency, queue depth,
transaction count -- anything a resource emits to Azure Monitor Metrics)
against a condition on a rolling window, and fires the referenced action
groups when the condition holds. Three condition families exist:

  - **static criteria** -- classic thresholds ("CPU average > 80%");
    multiple metrics may be combined in one rule (all must breach)
  - **dynamic criteria** -- machine-learning thresholds that learn the
    metric's normal band and alert on deviation; exactly one per rule
  - **web-test availability criteria** -- fires when an Application
    Insights availability test fails from N or more locations

A rule targets one or more `scopes`. Multi-resource rules (several
resources, a resource group, or a whole subscription in one rule) also
require target_resource_type and target_resource_location so Azure knows
which metric definition to evaluate.

Metric alerts are stateful by default (auto_mitigate): the alert fires
once when the condition starts holding and resolves itself when it stops.

## Example

```yaml
# Offline-plan test manifest. Exercises the dynamic-criteria path -- the
# only place the sensitivity map renders -- plus the GreaterOrLessThan
# operator, a dimension filter, the evaluation dials, an ignore-before
# instant, and action webhook properties. (The static path's map rows
# are exercised by the E2E scenarios.)
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorMetricAlert
metadata:
  name: test-metric-alert
  org: test-org
  env: dev
spec:
  resourceGroup:
    value: test-rg
  alertName: storage-transactions-anomaly
  scopes:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/testsa
  description: Transactions deviate from the learned band
  enabled: true
  autoMitigate: true
  severity: 2
  frequency: PT5M
  windowSize: PT15M
  dynamicCriteria:
    metricNamespace: Microsoft.Storage/storageAccounts
    metricName: Transactions
    aggregation: TOTAL
    operator: GREATER_OR_LESS_THAN
    alertSensitivity: MEDIUM
    evaluationTotalCount: 4
    evaluationFailureCount: 3
    ignoreDataBefore: "2026-01-15T00:00:00Z"
    dimensions:
      - name: ApiName
        operator: STARTS_WITH
        values:
          - Get
    skipMetricValidation: false
  actions:
    - actionGroupId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Insights/actionGroups/test-ag
      webhookProperties:
        source: planton
  tags:
    team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.alertName` | `string` | yes |  |  |
| `spec.scopes` | `[]string \| valueFrom` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.autoMitigate` | `bool` |  | `true` |  |
| `spec.severity` | `int32` |  | `3` |  |
| `spec.frequency` | `string` |  | `PT1M` |  |
| `spec.windowSize` | `string` |  | `PT5M` |  |
| `spec.staticCriteria` | `[]AzureMonitorMetricAlertStaticCriteria` |  |  |  |
| `spec.staticCriteria[].metricNamespace` | `string` | yes |  |  |
| `spec.staticCriteria[].metricName` | `string` | yes |  |  |
| `spec.staticCriteria[].aggregation` | `enum` |  |  |  |
| `spec.staticCriteria[].operator` | `enum` |  |  |  |
| `spec.staticCriteria[].threshold` | `double` |  |  |  |
| `spec.staticCriteria[].dimensions` | `[]AzureMonitorMetricAlertDimension` |  |  |  |
| `spec.staticCriteria[].dimensions[].name` | `string` | yes |  |  |
| `spec.staticCriteria[].dimensions[].operator` | `enum` |  |  |  |
| `spec.staticCriteria[].dimensions[].values` | `[]string` | yes |  |  |
| `spec.staticCriteria[].skipMetricValidation` | `bool` |  |  |  |
| `spec.dynamicCriteria` | `AzureMonitorMetricAlertDynamicCriteria` |  |  |  |
| `spec.dynamicCriteria.metricNamespace` | `string` | yes |  |  |
| `spec.dynamicCriteria.metricName` | `string` | yes |  |  |
| `spec.dynamicCriteria.aggregation` | `enum` |  |  |  |
| `spec.dynamicCriteria.operator` | `enum` |  |  |  |
| `spec.dynamicCriteria.alertSensitivity` | `enum` |  |  |  |
| `spec.dynamicCriteria.evaluationTotalCount` | `int32` |  | `4` |  |
| `spec.dynamicCriteria.evaluationFailureCount` | `int32` |  | `4` |  |
| `spec.dynamicCriteria.ignoreDataBefore` | `string` |  |  |  |
| `spec.dynamicCriteria.dimensions` | `[]AzureMonitorMetricAlertDimension` |  |  |  |
| `spec.dynamicCriteria.dimensions[].name` | `string` | yes |  |  |
| `spec.dynamicCriteria.dimensions[].operator` | `enum` |  |  |  |
| `spec.dynamicCriteria.dimensions[].values` | `[]string` | yes |  |  |
| `spec.dynamicCriteria.skipMetricValidation` | `bool` |  |  |  |
| `spec.webTestAvailabilityCriteria` | `AzureMonitorMetricAlertWebTestCriteria` |  |  |  |
| `spec.webTestAvailabilityCriteria.webTestId` | `string \| valueFrom` | yes |  | AzureApplicationInsightsStandardWebTest (`status.outputs.web_test_id`) |
| `spec.webTestAvailabilityCriteria.componentId` | `string \| valueFrom` | yes |  | AzureApplicationInsights (`status.outputs.application_insights_id`) |
| `spec.webTestAvailabilityCriteria.failedLocationCount` | `int32` |  |  |  |
| `spec.targetResourceType` | `string` |  |  |  |
| `spec.targetResourceLocation` | `string` |  |  |  |
| `spec.actions` | `[]AzureMonitorMetricAlertAction` |  |  |  |
| `spec.actions[].actionGroupId` | `string \| valueFrom` | yes |  | AzureMonitorActionGroup (`status.outputs.action_group_id`) |
| `spec.actions[].webhookProperties` | `map<string, string>` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group where the alert rule will be created.
Can be a literal string or a reference to an AzureResourceGroup output.
(Metric alert rules are global -- there is no region to choose.)

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.alertName

`string` · required

The name of the alert rule, unique within the resource group.

**ForceNew**: Changing this destroys and recreates the rule.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"260"}}

### spec.scopes

`[]string | valueFrom` · required

The resources whose metrics the rule evaluates -- each entry an ARM ID
of a resource, a resource group, or a subscription. There is no default
kind because any metric-emitting resource can be a scope: reference the
resource's `*_id` output explicitly with valueFrom (kind + fieldPath),
or pass a literal ARM ID. Multiple scopes (or a group/subscription
scope) require target_resource_type and target_resource_location.

Containment-exempt: an alert EVALUATES its scopes' metrics; it is
never deployed into one of them, and a rule over several scopes
could not live in all of them anyway. Without the exemption an
alert scoped to a storage account or a cluster by reference would
be drawn inside what it watches.

- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.description

`string`

A human-readable description delivered with every notification --
the place for runbook links and on-call context.

### spec.enabled

`bool` · optional (explicit presence)

Whether the rule evaluates at all. Disable during maintenance windows
instead of deleting the rule.
Default: true

- default: `true`

### spec.autoMitigate

`bool` · optional (explicit presence)

Whether the alert auto-resolves when the condition stops holding
(stateful behavior). Set false to fire on every evaluation the
condition holds -- noisy, but wanted when each occurrence must page.
Default: true (Azure's default)

- default: `true`

### spec.severity

`int32` · optional (explicit presence)

The alert severity: 0 (critical) through 4 (verbose). Severity drives
notification routing and on-call urgency conventions.
Default: 3 (informational -- raise deliberately for paging conditions)

- default: `3`
- rule: {"int32":{"lte":4,"gte":0}}

### spec.frequency

`string` · optional (explicit presence)

How often the rule evaluates, as an ISO 8601 duration. One of PT1M,
PT5M, PT15M, PT30M, PT1H. Must not exceed window_size.
Default: PT1M (evaluate every minute)

- default: `PT1M`
- rule: {"string":{"in":["PT1M","PT5M","PT15M","PT30M","PT1H"]}}

### spec.windowSize

`string` · optional (explicit presence)

The rolling window each evaluation aggregates over, as an ISO 8601
duration. One of PT1M, PT5M, PT15M, PT30M, PT1H, PT6H, PT12H, P1D.
Must be at least the evaluation frequency.
Default: PT5M (aggregate the last five minutes)

- default: `PT5M`
- rule: {"string":{"in":["PT1M","PT5M","PT15M","PT30M","PT1H","PT6H","PT12H","P1D"]}}

### spec.staticCriteria

`[]AzureMonitorMetricAlertStaticCriteria`

Static threshold conditions -- the classic "metric crosses a value"
family. Multiple criteria combine with AND (all must breach for the
alert to fire). Exactly one condition family per rule.

### spec.staticCriteria[].metricNamespace

`string` · required

The metric namespace (for example "Microsoft.Storage/storageAccounts" --
usually the resource type; custom metrics use their custom namespace).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.staticCriteria[].metricName

`string` · required

The metric to evaluate (for example "Transactions", "Percentage CPU").
Metric names are defined per resource type by Azure Monitor.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.staticCriteria[].aggregation

`enum`

How samples in the window aggregate before comparison.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_monitor_metric_alert_aggregation_unspecified` -- Not specified -- invalid; choose an explicit aggregation.
- `AVERAGE` -- The arithmetic mean of samples in the window.
- `COUNT` -- The number of samples in the window.
- `MINIMUM` -- The smallest sample in the window.
- `MAXIMUM` -- The largest sample in the window.
- `TOTAL` -- The sum of samples in the window.

### spec.staticCriteria[].operator

`enum`

The comparison between the aggregated value and the threshold. Static
criteria compare in one direction: EQUALS, GREATER_THAN,
GREATER_THAN_OR_EQUAL, LESS_THAN, or LESS_THAN_OR_EQUAL
(GREATER_OR_LESS_THAN belongs to dynamic criteria).

- rule: static criteria compare in one direction -- use EQUALS, GREATER_THAN, GREATER_THAN_OR_EQUAL, LESS_THAN, or LESS_THAN_OR_EQUAL (GREATER_OR_LESS_THAN is a dynamic-criteria operator)
- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_monitor_metric_alert_operator_unspecified` -- Not specified -- invalid; choose an explicit operator.
- `EQUALS` -- The aggregated value equals the threshold (static only).
- `GREATER_THAN` -- The aggregated value exceeds the threshold / learned band.
- `GREATER_THAN_OR_EQUAL` -- The aggregated value is at least the threshold (static only).
- `LESS_THAN` -- The aggregated value is below the threshold / learned band.
- `LESS_THAN_OR_EQUAL` -- The aggregated value is at most the threshold (static only).
- `GREATER_OR_LESS_THAN` -- The value deviates in either direction from the learned band (dynamic only).

### spec.staticCriteria[].threshold

`double`

The threshold the aggregated value is compared against. Zero is a
meaningful threshold (for example "any failed request").

### spec.staticCriteria[].dimensions

`[]AzureMonitorMetricAlertDimension`

Dimension filters -- restrict the evaluation to specific dimension
values (for example only the "GetBlob" API name) or split it across
them (each matching dimension value alerts independently).

### spec.staticCriteria[].dimensions[].name

`string` · required

The dimension name (for example "ApiName", "StatusCode"). Dimensions
are defined per metric by Azure Monitor.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.staticCriteria[].dimensions[].operator

`enum`

How the values list applies: INCLUDE evaluates only matching values
(each independently), EXCLUDE evaluates everything else, STARTS_WITH
matches by prefix.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_monitor_metric_alert_dimension_operator_unspecified` -- Not specified -- invalid; choose an explicit operator.
- `INCLUDE` -- Evaluate only the listed values, each independently.
- `EXCLUDE` -- Evaluate every value except the listed ones.
- `STARTS_WITH` -- Evaluate values matching the listed prefixes.

### spec.staticCriteria[].dimensions[].values

`[]string` · required

The dimension values the operator applies to. "*" with INCLUDE splits
the alert across every value of the dimension.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.staticCriteria[].skipMetricValidation

`bool`

Whether to skip Azure's validation that the metric exists on the scoped
resources. Needed when the metric appears only after the resource
starts emitting (custom metrics).

### spec.dynamicCriteria

`AzureMonitorMetricAlertDynamicCriteria`

A dynamic (machine-learning) threshold condition -- Azure learns the
metric's normal band and alerts on deviation. Exactly one dynamic
criterion per rule; exactly one condition family per rule.

### spec.dynamicCriteria.metricNamespace

`string` · required

The metric namespace (usually the resource type).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dynamicCriteria.metricName

`string` · required

The metric to evaluate.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dynamicCriteria.aggregation

`enum`

How samples in the window aggregate before comparison.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_monitor_metric_alert_aggregation_unspecified` -- Not specified -- invalid; choose an explicit aggregation.
- `AVERAGE` -- The arithmetic mean of samples in the window.
- `COUNT` -- The number of samples in the window.
- `MINIMUM` -- The smallest sample in the window.
- `MAXIMUM` -- The largest sample in the window.
- `TOTAL` -- The sum of samples in the window.

### spec.dynamicCriteria.operator

`enum`

Which deviations from the learned band alert: GREATER_THAN (spikes),
LESS_THAN (drops), or GREATER_OR_LESS_THAN (both directions).

- rule: dynamic criteria alert on deviation from the learned band -- use GREATER_THAN, LESS_THAN, or GREATER_OR_LESS_THAN
- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_monitor_metric_alert_operator_unspecified` -- Not specified -- invalid; choose an explicit operator.
- `EQUALS` -- The aggregated value equals the threshold (static only).
- `GREATER_THAN` -- The aggregated value exceeds the threshold / learned band.
- `GREATER_THAN_OR_EQUAL` -- The aggregated value is at least the threshold (static only).
- `LESS_THAN` -- The aggregated value is below the threshold / learned band.
- `LESS_THAN_OR_EQUAL` -- The aggregated value is at most the threshold (static only).
- `GREATER_OR_LESS_THAN` -- The value deviates in either direction from the learned band (dynamic only).

### spec.dynamicCriteria.alertSensitivity

`enum`

How tightly the learned band hugs the metric: HIGH sensitivity alerts
on small deviations (more alerts), LOW only on large ones.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_monitor_metric_alert_sensitivity_unspecified` -- Not specified -- invalid; choose an explicit sensitivity.
- `LOW` -- Alert only on large deviations -- the quietest setting.
- `MEDIUM` -- Balanced deviation detection.
- `HIGH` -- Alert on small deviations -- the most sensitive setting.

### spec.dynamicCriteria.evaluationTotalCount

`int32` · optional (explicit presence)

The number of recent evaluation periods examined per evaluation
(the lookback, 1-6 in practice).
Default: 4

- default: `4`
- rule: {"int32":{"gte":1}}

### spec.dynamicCriteria.evaluationFailureCount

`int32` · optional (explicit presence)

How many of the examined periods must deviate for the alert to fire.
Must not exceed evaluation_total_count (Azure rejects the rule
otherwise).
Default: 4

- default: `4`
- rule: {"int32":{"gte":1}}

### spec.dynamicCriteria.ignoreDataBefore

`string`

An RFC 3339 timestamp before which history is ignored when learning the
band -- use after a known regime change (migration, launch) so the old
normal stops informing the threshold.

- rule: ignore_data_before must be an RFC 3339 timestamp, e.g. 2026-01-15T00:00:00Z

### spec.dynamicCriteria.dimensions

`[]AzureMonitorMetricAlertDimension`

Dimension filters -- restrict or split the evaluation across dimension
values.

### spec.dynamicCriteria.dimensions[].name

`string` · required

The dimension name (for example "ApiName", "StatusCode"). Dimensions
are defined per metric by Azure Monitor.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dynamicCriteria.dimensions[].operator

`enum`

How the values list applies: INCLUDE evaluates only matching values
(each independently), EXCLUDE evaluates everything else, STARTS_WITH
matches by prefix.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_monitor_metric_alert_dimension_operator_unspecified` -- Not specified -- invalid; choose an explicit operator.
- `INCLUDE` -- Evaluate only the listed values, each independently.
- `EXCLUDE` -- Evaluate every value except the listed ones.
- `STARTS_WITH` -- Evaluate values matching the listed prefixes.

### spec.dynamicCriteria.dimensions[].values

`[]string` · required

The dimension values the operator applies to. "*" with INCLUDE splits
the alert across every value of the dimension.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dynamicCriteria.skipMetricValidation

`bool`

Whether to skip Azure's validation that the metric exists on the scoped
resources.

### spec.webTestAvailabilityCriteria

`AzureMonitorMetricAlertWebTestCriteria`

An Application Insights web-test availability condition -- fires when
the referenced availability test fails from the configured number of
locations. Exactly one condition family per rule.

### spec.webTestAvailabilityCriteria.webTestId

`string | valueFrom` · required

The Application Insights availability (web) test whose failures the
rule watches. Defaults to referencing an
AzureApplicationInsightsStandardWebTest's web_test_id output.

- references: AzureApplicationInsightsStandardWebTest (`status.outputs.web_test_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureApplicationInsightsStandardWebTest, name: <that resource's name>, fieldPath: status.outputs.web_test_id}} -- a bare string does not parse

### spec.webTestAvailabilityCriteria.componentId

`string | valueFrom` · required

The Application Insights resource the web test belongs to. Can be a
literal ARM ID or a reference to an AzureApplicationInsights output.

- references: AzureApplicationInsights (`status.outputs.application_insights_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureApplicationInsights, name: <that resource's name>, fieldPath: status.outputs.application_insights_id}} -- a bare string does not parse

### spec.webTestAvailabilityCriteria.failedLocationCount

`int32`

How many test locations must fail simultaneously for the alert to fire.
Set below the test's location count to tolerate single-location blips
(a common choice: locations minus two).

- rule: {"int32":{"gte":1}}

### spec.targetResourceType

`string`

The resource type the rule evaluates (for example
"Microsoft.Storage/storageAccounts"). Required when scopes span
multiple resources or target a resource group or subscription --
Azure needs it to resolve the metric definition. Leave empty for a
single-resource scope.

### spec.targetResourceLocation

`string`

The region of the resources the rule evaluates (for example "eastus").
Required alongside target_resource_type for multi-resource, resource
group, or subscription scopes. Leave empty for a single-resource scope.

### spec.actions

`[]AzureMonitorMetricAlertAction`

The action groups to fire when the alert triggers or resolves.

### spec.actions[].actionGroupId

`string | valueFrom` · required

The action group to notify. Can be a literal ARM ID or a reference to
an AzureMonitorActionGroup output.

- references: AzureMonitorActionGroup (`status.outputs.action_group_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMonitorActionGroup, name: <that resource's name>, fieldPath: status.outputs.action_group_id}} -- a bare string does not parse

### spec.actions[].webhookProperties

`map<string, string>`

Custom properties merged into the webhook payload delivered to the
action group's webhook receivers.

### spec.tags

`map<string, string>`

Tags to apply to the alert rule, merged over the Planton-derived
metadata tags (user values win on key conflicts).

## Validation Rules

- `metric_alert_exactly_one_criteria_family`: configure exactly one condition family: static_criteria (one or more thresholds), dynamic_criteria, or web_test_availability_criteria

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMonitorMetricAlert, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.metric_alert_id` | `string` | The Azure Resource Manager ID of the metric alert rule. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Insights/metricAlerts/{name} |
| `status.outputs.metric_alert_name` | `string` | The name of the metric alert rule. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.webTestAvailabilityCriteria.webTestId` | AzureApplicationInsightsStandardWebTest | `status.outputs.web_test_id` |
| `spec.webTestAvailabilityCriteria.componentId` | AzureApplicationInsights | `status.outputs.application_insights_id` |
| `spec.actions[].actionGroupId` | AzureMonitorActionGroup | `status.outputs.action_group_id` |

## See Also

- [Overview](../README.md)
