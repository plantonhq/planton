# AzureMonitorAutoscaleSetting

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureMonitorAutoscaleSettingSpec** defines an Azure Monitor autoscale
setting -- the rule book that automatically adds and removes instances
of ONE scalable target (a Virtual Machine Scale Set, an App Service
plan, and other capacity-bearing resources) based on metrics and
schedules.

A setting holds up to 20 **profiles**. Exactly one profile is in effect
at any moment, chosen by Azure in this precedence order: a fixed_date
profile whose window covers "now" wins, then a recurrence profile whose
schedule matches, then the default profile (the one with neither).
Each profile carries a capacity envelope (minimum/maximum/default
instance counts) and up to 10 metric-driven rules ("CPU average over
10 minutes > 75% -> add 2 instances, then cool down 5 minutes").

**One setting per target**: Azure allows a single autoscale setting per
target resource; a second setting on the same target fails at deploy
time. The setting must live in the same region as its target.

**Capacity is bounded by quota**: the maximum instance count is also
limited by the subscription's core quota at deploy time -- the 0-1000
bounds here mirror the provider's static validation only.

## Example

```yaml
# Offline-plan test manifest. Exercises the DEEP shape the live smoke
# does not: three profiles (metric rules with dimensions, namespace, and
# per-instance division; a timezone'd recurrence; a fixed-date window),
# predictive autoscale, and both notification channels with webhook
# properties. (The live scenario proves the everyday shape: default
# profile + rule + recurrence + email.)
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorAutoscaleSetting
metadata:
  name: test-autoscale-setting
  org: test-org
  env: dev
spec:
  resourceGroup:
    value: test-rg
  name: web-pool-autoscale
  region: eastus
  targetResourceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachineScaleSets/web-pool
  enabled: true
  predictive:
    scaleMode: ForecastOnly
    lookAheadTime: PT10M
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
              value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachineScaleSets/web-pool
            timeGrain: PT1M
            statistic: Average
            timeWindow: PT10M
            timeAggregation: Average
            operator: GreaterThan
            threshold: 75
            metricNamespace: microsoft.compute/virtualmachinescalesets
            divideByInstanceCount: true
            dimensions:
              - name: VMName
                operator: NotEquals
                values:
                  - canary-0
          scaleAction:
            direction: Increase
            type: PercentChangeCount
            value: 20
            cooldown: PT5M
        - metricTrigger:
            metricName: Percentage CPU
            metricResourceId:
              value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachineScaleSets/web-pool
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
    - name: weekday-business-hours
      capacity:
        minimum: 4
        maximum: 12
        default: 4
      recurrence:
        timezone: Eastern Standard Time
        days: [Monday, Tuesday, Wednesday, Thursday, Friday]
        hour: 8
        minute: 30
    - name: launch-window
      capacity:
        minimum: 8
        maximum: 20
        default: 8
      fixedDate:
        timezone: Pacific Standard Time
        start: "2026-11-27T00:00:00Z"
        end: "2026-11-30T00:00:00Z"
  notification:
    email:
      # Explicit recipients are the only email wiring ARM still accepts --
      # the classic admin/co-admin flags were retired in April 2024.
      customEmails:
        - oncall@example.com
    webhooks:
      - serviceUri: https://hooks.example.com/scale-events
        properties:
          channel: ops
  tags:
    team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.targetResourceId` | `string \| valueFrom` | yes |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.predictive` | `AzureMonitorAutoscaleSettingPredictive` |  |  |  |
| `spec.predictive.scaleMode` | `string` | yes |  |  |
| `spec.predictive.lookAheadTime` | `string` |  |  |  |
| `spec.profiles` | `[]AzureMonitorAutoscaleSettingProfile` | yes |  |  |
| `spec.profiles[].name` | `string` | yes |  |  |
| `spec.profiles[].capacity` | `AzureMonitorAutoscaleSettingCapacity` | yes |  |  |
| `spec.profiles[].capacity.minimum` | `int32` | yes |  |  |
| `spec.profiles[].capacity.maximum` | `int32` | yes |  |  |
| `spec.profiles[].capacity.default` | `int32` | yes |  |  |
| `spec.profiles[].rules` | `[]AzureMonitorAutoscaleSettingRule` |  |  |  |
| `spec.profiles[].rules[].metricTrigger` | `AzureMonitorAutoscaleSettingMetricTrigger` | yes |  |  |
| `spec.profiles[].rules[].metricTrigger.metricName` | `string` | yes |  |  |
| `spec.profiles[].rules[].metricTrigger.metricResourceId` | `string \| valueFrom` | yes |  |  |
| `spec.profiles[].rules[].metricTrigger.timeGrain` | `string` | yes |  |  |
| `spec.profiles[].rules[].metricTrigger.statistic` | `string` | yes |  |  |
| `spec.profiles[].rules[].metricTrigger.timeWindow` | `string` | yes |  |  |
| `spec.profiles[].rules[].metricTrigger.timeAggregation` | `string` | yes |  |  |
| `spec.profiles[].rules[].metricTrigger.operator` | `string` | yes |  |  |
| `spec.profiles[].rules[].metricTrigger.threshold` | `double` |  |  |  |
| `spec.profiles[].rules[].metricTrigger.metricNamespace` | `string` |  |  |  |
| `spec.profiles[].rules[].metricTrigger.divideByInstanceCount` | `bool` |  |  |  |
| `spec.profiles[].rules[].metricTrigger.dimensions` | `[]AzureMonitorAutoscaleSettingDimension` |  |  |  |
| `spec.profiles[].rules[].metricTrigger.dimensions[].name` | `string` | yes |  |  |
| `spec.profiles[].rules[].metricTrigger.dimensions[].operator` | `string` | yes |  |  |
| `spec.profiles[].rules[].metricTrigger.dimensions[].values` | `[]string` | yes |  |  |
| `spec.profiles[].rules[].scaleAction` | `AzureMonitorAutoscaleSettingScaleAction` | yes |  |  |
| `spec.profiles[].rules[].scaleAction.direction` | `string` | yes |  |  |
| `spec.profiles[].rules[].scaleAction.type` | `string` | yes |  |  |
| `spec.profiles[].rules[].scaleAction.value` | `int32` | yes |  |  |
| `spec.profiles[].rules[].scaleAction.cooldown` | `string` | yes |  |  |
| `spec.profiles[].fixedDate` | `AzureMonitorAutoscaleSettingFixedDate` |  |  |  |
| `spec.profiles[].fixedDate.timezone` | `string` |  | `UTC` |  |
| `spec.profiles[].fixedDate.start` | `string` | yes |  |  |
| `spec.profiles[].fixedDate.end` | `string` | yes |  |  |
| `spec.profiles[].recurrence` | `AzureMonitorAutoscaleSettingRecurrence` |  |  |  |
| `spec.profiles[].recurrence.timezone` | `string` |  | `UTC` |  |
| `spec.profiles[].recurrence.days` | `[]string` | yes |  |  |
| `spec.profiles[].recurrence.hour` | `int32` | yes |  |  |
| `spec.profiles[].recurrence.minute` | `int32` | yes |  |  |
| `spec.notification` | `AzureMonitorAutoscaleSettingNotification` |  |  |  |
| `spec.notification.email` | `AzureMonitorAutoscaleSettingNotificationEmail` |  |  |  |
| `spec.notification.email.customEmails` | `[]string` |  |  |  |
| `spec.notification.webhooks` | `[]AzureMonitorAutoscaleSettingNotificationWebhook` |  |  |  |
| `spec.notification.webhooks[].serviceUri` | `string` | yes |  |  |
| `spec.notification.webhooks[].properties` | `map<string, string>` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the autoscale setting lives in. Can be a
literal string or a reference to an AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the setting.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The name of the autoscale setting, unique within the resource group.

**ForceNew**: changing this destroys and recreates the setting.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.region

`string` · required

The Azure region the setting is created in, e.g. "eastus". Must match
the region of the target resource (Azure evaluates autoscale in the
target's region).

**ForceNew**: changing this destroys and recreates the setting.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.targetResourceId

`string | valueFrom` · required

The scalable resource this setting controls, by ARM resource ID --
a Virtual Machine Scale Set, an App Service plan (Standard tier or
above), or any other resource type Azure Monitor autoscale supports.
There is no default kind because many kinds can be the target:
reference the resource's `*_id` output explicitly with valueFrom
(kind + fieldPath), or pass a literal ARM ID. Azure allows ONE
autoscale setting per target.

Containment-exempt: the setting DRIVES its target's capacity from
outside; it is an extension resource that hangs on the scale set or
plan, never a tenant of it. Without the exemption a setting wired
to a container target by reference would be drawn inside it.

**ForceNew**: changing this destroys and recreates the setting.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.enabled

`bool` · optional (explicit presence)

Whether autoscale actively evaluates and acts. Unspecified applies
true -- a setting exists to scale; set false to freeze the instance
count while keeping the whole rule book in place.

- default: `true`

### spec.predictive

`AzureMonitorAutoscaleSettingPredictive`

Predictive autoscale (VM Scale Sets only): Azure forecasts CPU load
from history and scales out AHEAD of the predicted demand. OMIT this
block to leave predictive autoscale disabled -- there is no explicit
"Disabled" mode; absence is how the API expresses it.

### spec.predictive.scaleMode

`string` · required

The predictive mode: "ForecastOnly" produces the forecast for charts
without acting on it (the safe first step); "Enabled" also scales on
the forecast. There is no "Disabled" value -- omit the whole
predictive block to disable.

- rule: {"required":true,"string":{"in":["Enabled","ForecastOnly"]}}

### spec.predictive.lookAheadTime

`string`

How far ahead of the forecast instances are launched, as an ISO 8601
duration between PT1M and PT1H (for example "PT10M" to have capacity
ready ten minutes early). Omit to scale exactly at the predicted
time.

- rule: look_ahead_time must be an ISO 8601 duration between PT1M and PT1H, e.g. PT10M

### spec.profiles

`[]AzureMonitorAutoscaleSettingProfile` · required

The scaling profiles (1-20). Exactly one is in effect at any moment:
a matching fixed_date profile wins over a matching recurrence
profile, which wins over the default profile (no schedule). Ship at
least the default profile; add scheduled profiles for known load
patterns (business hours, weekends, launch events).

- rule: {"repeated":{"minItems":"1","maxItems":"20"}}
- rule: a profile is default (no schedule), fixed-date, or recurring -- set at most one of fixed_date and recurrence

### spec.profiles[].name

`string` · required

The profile's name (shown in the portal's autoscale history --
choose something an operator recognizes, e.g. "weekday-business-hours").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.profiles[].capacity

`AzureMonitorAutoscaleSettingCapacity` · required

The instance-count envelope this profile scales within.

- rule: {"required":true}

### spec.profiles[].capacity.minimum

`int32` · required · optional (explicit presence)

The floor: autoscale never goes below this many instances.

- rule: {"required":true,"int32":{"lte":1000,"gte":0}}

### spec.profiles[].capacity.maximum

`int32` · required · optional (explicit presence)

The ceiling: autoscale never goes above this many instances.

- rule: {"required":true,"int32":{"lte":1000,"gte":0}}

### spec.profiles[].capacity.default

`int32` · required · optional (explicit presence)

The fallback instance count Azure applies when metrics are
unavailable for evaluation -- and only if the current count is
BELOW it (metrics loss never scales you down).

- rule: {"required":true,"int32":{"lte":1000,"gte":0}}

### spec.profiles[].rules

`[]AzureMonitorAutoscaleSettingRule`

The metric-driven rules (0-10) that move the instance count inside
the capacity envelope. A profile with no rules pins the count to the
capacity default -- the standard shape for scheduled profiles that
simply set a fixed size for a time window.

- rule: {"repeated":{"maxItems":"10"}}

### spec.profiles[].rules[].metricTrigger

`AzureMonitorAutoscaleSettingMetricTrigger` · required

The metric condition that triggers the rule.

- rule: {"required":true}

### spec.profiles[].rules[].metricTrigger.metricName

`string` · required

The metric to evaluate (for example "Percentage CPU" on a VM Scale
Set, "CpuPercentage" on an App Service plan). Metric names are
defined per resource type by Azure Monitor.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.profiles[].rules[].metricTrigger.metricResourceId

`string | valueFrom` · required

The resource whose metric is evaluated, by ARM resource ID --
usually the scale target itself (reference the same output as
target_resource_id), but any metric-emitting resource works (e.g.
scale a worker scale set on a Service Bus queue's depth). No
default kind for the same reason as target_resource_id: reference
the resource's `*_id` output explicitly with valueFrom, or pass a
literal ARM ID. Access, not placement, like target_resource_id --
the setting reads this resource's metric, nothing more.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.profiles[].rules[].metricTrigger.timeGrain

`string` · required

The granularity the metric is sampled at, as an ISO 8601 duration
(commonly "PT1M"). Must be one of the granularities the metric
supports -- Azure rejects unsupported values at deploy time.

- rule: time_grain must be an ISO 8601 duration, e.g. PT1M
- rule: {"required":true}

### spec.profiles[].rules[].metricTrigger.statistic

`string` · required

How samples within one time grain combine: "Average", "Max", "Min",
or "Sum".

- rule: {"required":true,"string":{"in":["Average","Max","Min","Sum"]}}

### spec.profiles[].rules[].metricTrigger.timeWindow

`string` · required

The rolling window the rule looks back over, as an ISO 8601 duration
(commonly "PT5M" to "PT30M"). Longer windows smooth spikes; shorter
windows react faster.

- rule: time_window must be an ISO 8601 duration, e.g. PT10M
- rule: {"required":true}

### spec.profiles[].rules[].metricTrigger.timeAggregation

`string` · required

How the per-grain statistics aggregate across the whole window
before comparison: "Average", "Count", "Maximum", "Minimum",
"Total", or "Last".

- rule: {"required":true,"string":{"in":["Average","Count","Maximum","Minimum","Total","Last"]}}

### spec.profiles[].rules[].metricTrigger.operator

`string` · required

The comparison between the aggregated value and the threshold:
"Equals", "NotEquals", "GreaterThan", "GreaterThanOrEqual",
"LessThan", or "LessThanOrEqual".

- rule: {"required":true,"string":{"in":["Equals","NotEquals","GreaterThan","GreaterThanOrEqual","LessThan","LessThanOrEqual"]}}

### spec.profiles[].rules[].metricTrigger.threshold

`double`

The threshold the aggregated metric is compared against. Zero is a
meaningful threshold (for example "queue depth greater than 0").

### spec.profiles[].rules[].metricTrigger.metricNamespace

`string`

The metric namespace, when the metric lives outside the resource
type's default namespace (custom metrics; e.g.
"microsoft.compute/virtualmachinescalesets" is implied for a scale
set's own platform metrics and can be omitted).

### spec.profiles[].rules[].metricTrigger.divideByInstanceCount

`bool`

Whether the metric is divided by the current instance count before
comparison -- useful for per-instance semantics on aggregate metrics
(e.g. total requests per instance).

### spec.profiles[].rules[].metricTrigger.dimensions

`[]AzureMonitorAutoscaleSettingDimension`

Dimension filters -- restrict the evaluation to specific dimension
values (for example only one storage queue's depth, only one
node type's CPU).

### spec.profiles[].rules[].metricTrigger.dimensions[].name

`string` · required

The dimension name (for example "AppName", "Instance"). Dimensions
are defined per metric by Azure Monitor.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.profiles[].rules[].metricTrigger.dimensions[].operator

`string` · required

How the values list applies: "Equals" evaluates only matching
values, "NotEquals" evaluates everything else.

- rule: {"required":true,"string":{"in":["Equals","NotEquals"]}}

### spec.profiles[].rules[].metricTrigger.dimensions[].values

`[]string` · required

The dimension values the operator applies to.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.profiles[].rules[].scaleAction

`AzureMonitorAutoscaleSettingScaleAction` · required

The scale action taken when the condition holds.

- rule: {"required":true}

### spec.profiles[].rules[].scaleAction.direction

`string` · required

Which way the instance count moves: "Increase" or "Decrease".

- rule: {"required":true,"string":{"in":["Increase","Decrease"]}}

### spec.profiles[].rules[].scaleAction.type

`string` · required

How `value` is interpreted: "ChangeCount" adds/removes that many
instances, "PercentChangeCount" changes by that percentage of the
current count, "ExactCount" jumps straight to that count, and
"ServiceAllowedNextValue" steps to the next value the target service
allows (for services with discrete capacity steps).

- rule: {"required":true,"string":{"in":["ChangeCount","ExactCount","PercentChangeCount","ServiceAllowedNextValue"]}}

### spec.profiles[].rules[].scaleAction.value

`int32` · required · optional (explicit presence)

The magnitude of the action, interpreted per `type`. Zero is legal
(a rule that intentionally holds the count).

- rule: {"required":true,"int32":{"gte":0}}

### spec.profiles[].rules[].scaleAction.cooldown

`string` · required

How long autoscale waits after this action before evaluating rules
again, as an ISO 8601 duration between one minute and one week
(commonly "PT5M"). Cooldowns prevent flapping -- scale-in rules
usually carry longer ones than scale-out rules.

- rule: cooldown must be an ISO 8601 duration, e.g. PT5M
- rule: {"required":true}

### spec.profiles[].fixedDate

`AzureMonitorAutoscaleSettingFixedDate`

Makes this profile active for ONE fixed calendar window (e.g. a
launch day). At most one of fixed_date and recurrence may be set;
a profile with neither is the default profile.

### spec.profiles[].fixedDate.timezone

`string` · optional (explicit presence)

The IANA-style Azure time zone `start` and `end` are expressed in,
from Azure's fixed vocabulary (e.g. "Pacific Standard Time").
Unspecified applies "UTC".

- default: `UTC`
- rule: {"string":{"in":["Dateline Standard Time","UTC-11","Hawaiian Standard Time","Alaskan Standard Time","Pacific Standard Time (Mexico)","Pacific Standard Time","US Mountain Standard Time","Mountain Standard Time (Mexico)","Mountain Standard Time","Central America Standard Time","Central Standard Time","Central Standard Time (Mexico)","Canada Central Standard Time","SA Pacific Standard Time","Eastern Standard Time","US Eastern Standard Time","Venezuela Standard Time","Paraguay Standard Time","Atlantic Standard Time","Central Brazilian Standard Time","SA Western Standard Time","Pacific SA Standard Time","Newfoundland Standard Time","E. South America Standard Time","Argentina Standard Time","SA Eastern Standard Time","Greenland Standard Time","Montevideo Standard Time","Bahia Standard Time","UTC-02","Mid-Atlantic Standard Time","Azores Standard Time","Cape Verde Standard Time","Morocco Standard Time","UTC","GMT Standard Time","Greenwich Standard Time","W. Europe Standard Time","Central Europe Standard Time","Romance Standard Time","Central European Standard Time","W. Central Africa Standard Time","Namibia Standard Time","Jordan Standard Time","GTB Standard Time","Middle East Standard Time","Egypt Standard Time","Syria Standard Time","E. Europe Standard Time","South Africa Standard Time","FLE Standard Time","Turkey Standard Time","Israel Standard Time","Kaliningrad Standard Time","Libya Standard Time","Arabic Standard Time","Arab Standard Time","Belarus Standard Time","Russian Standard Time","E. Africa Standard Time","Iran Standard Time","Arabian Standard Time","Azerbaijan Standard Time","Russia Time Zone 3","Mauritius Standard Time","Georgian Standard Time","Caucasus Standard Time","Afghanistan Standard Time","West Asia Standard Time","Ekaterinburg Standard Time","Pakistan Standard Time","India Standard Time","Sri Lanka Standard Time","Nepal Standard Time","Central Asia Standard Time","Bangladesh Standard Time","N. Central Asia Standard Time","Myanmar Standard Time","SE Asia Standard Time","North Asia Standard Time","China Standard Time","North Asia East Standard Time","Singapore Standard Time","W. Australia Standard Time","Taipei Standard Time","Ulaanbaatar Standard Time","Tokyo Standard Time","Korea Standard Time","Yakutsk Standard Time","Cen. Australia Standard Time","AUS Central Standard Time","E. Australia Standard Time","AUS Eastern Standard Time","West Pacific Standard Time","Tasmania Standard Time","Magadan Standard Time","Vladivostok Standard Time","Russia Time Zone 10","Central Pacific Standard Time","Russia Time Zone 11","New Zealand Standard Time","UTC+12","Fiji Standard Time","Kamchatka Standard Time","Tonga Standard Time","Samoa Standard Time","Line Islands Standard Time"]}}

### spec.profiles[].fixedDate.start

`string` · required

When the profile starts, as an RFC 3339 timestamp
(e.g. "2026-11-27T00:00:00Z").

- rule: start must be an RFC 3339 timestamp, e.g. 2026-11-27T00:00:00Z
- rule: {"required":true}

### spec.profiles[].fixedDate.end

`string` · required

When the profile ends, as an RFC 3339 timestamp. After the end time
the default (or recurring) profile takes over again.

- rule: end must be an RFC 3339 timestamp, e.g. 2026-11-30T00:00:00Z
- rule: {"required":true}

### spec.profiles[].recurrence

`AzureMonitorAutoscaleSettingRecurrence`

Makes this profile active on a weekly schedule (e.g. weekdays at
08:00). The profile stays in effect until another profile's schedule
starts -- schedule the "return to normal" profile explicitly. At
most one of fixed_date and recurrence may be set.

### spec.profiles[].recurrence.timezone

`string` · optional (explicit presence)

The IANA-style Azure time zone the schedule is evaluated in, from
Azure's fixed vocabulary (e.g. "Eastern Standard Time").
Unspecified applies "UTC".

- default: `UTC`
- rule: {"string":{"in":["Dateline Standard Time","UTC-11","Hawaiian Standard Time","Alaskan Standard Time","Pacific Standard Time (Mexico)","Pacific Standard Time","US Mountain Standard Time","Mountain Standard Time (Mexico)","Mountain Standard Time","Central America Standard Time","Central Standard Time","Central Standard Time (Mexico)","Canada Central Standard Time","SA Pacific Standard Time","Eastern Standard Time","US Eastern Standard Time","Venezuela Standard Time","Paraguay Standard Time","Atlantic Standard Time","Central Brazilian Standard Time","SA Western Standard Time","Pacific SA Standard Time","Newfoundland Standard Time","E. South America Standard Time","Argentina Standard Time","SA Eastern Standard Time","Greenland Standard Time","Montevideo Standard Time","Bahia Standard Time","UTC-02","Mid-Atlantic Standard Time","Azores Standard Time","Cape Verde Standard Time","Morocco Standard Time","UTC","GMT Standard Time","Greenwich Standard Time","W. Europe Standard Time","Central Europe Standard Time","Romance Standard Time","Central European Standard Time","W. Central Africa Standard Time","Namibia Standard Time","Jordan Standard Time","GTB Standard Time","Middle East Standard Time","Egypt Standard Time","Syria Standard Time","E. Europe Standard Time","South Africa Standard Time","FLE Standard Time","Turkey Standard Time","Israel Standard Time","Kaliningrad Standard Time","Libya Standard Time","Arabic Standard Time","Arab Standard Time","Belarus Standard Time","Russian Standard Time","E. Africa Standard Time","Iran Standard Time","Arabian Standard Time","Azerbaijan Standard Time","Russia Time Zone 3","Mauritius Standard Time","Georgian Standard Time","Caucasus Standard Time","Afghanistan Standard Time","West Asia Standard Time","Ekaterinburg Standard Time","Pakistan Standard Time","India Standard Time","Sri Lanka Standard Time","Nepal Standard Time","Central Asia Standard Time","Bangladesh Standard Time","N. Central Asia Standard Time","Myanmar Standard Time","SE Asia Standard Time","North Asia Standard Time","China Standard Time","North Asia East Standard Time","Singapore Standard Time","W. Australia Standard Time","Taipei Standard Time","Ulaanbaatar Standard Time","Tokyo Standard Time","Korea Standard Time","Yakutsk Standard Time","Cen. Australia Standard Time","AUS Central Standard Time","E. Australia Standard Time","AUS Eastern Standard Time","West Pacific Standard Time","Tasmania Standard Time","Magadan Standard Time","Vladivostok Standard Time","Russia Time Zone 10","Central Pacific Standard Time","Russia Time Zone 11","New Zealand Standard Time","UTC+12","Fiji Standard Time","Kamchatka Standard Time","Tonga Standard Time","Samoa Standard Time","Line Islands Standard Time"]}}

### spec.profiles[].recurrence.days

`[]string` · required

The days of the week the profile starts on.

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["Monday","Tuesday","Wednesday","Thursday","Friday","Saturday","Sunday"]}}}}

### spec.profiles[].recurrence.hour

`int32` · required · optional (explicit presence)

The hour of day (0-23) the profile starts at, in the schedule's
timezone.

- rule: {"required":true,"int32":{"lte":23,"gte":0}}

### spec.profiles[].recurrence.minute

`int32` · required · optional (explicit presence)

The minute past the hour (0-59) the profile starts at.

- rule: {"required":true,"int32":{"lte":59,"gte":0}}

### spec.notification

`AzureMonitorAutoscaleSettingNotification`

Scale-event notifications: email to administrators/addresses and/or
webhooks fired on every scale action. Omit when no notification is
wanted.

- rule: configure at least one notification channel: email and/or webhooks

### spec.notification.email

`AzureMonitorAutoscaleSettingNotificationEmail`

Email notification settings. At least one of email and webhooks must
be configured.

- rule: email notifications need at least one address in custom_emails (Azure retired the subscription-administrator email flags in April 2024)

### spec.notification.email.customEmails

`[]string`

Email addresses notified on every scale action.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.notification.webhooks

`[]AzureMonitorAutoscaleSettingNotificationWebhook`

Webhooks fired on every scale action (wire these to chat, paging,
or automation endpoints).

### spec.notification.webhooks[].serviceUri

`string` · required

The URI Azure POSTs the scale-event payload to.

- rule: the webhook service_uri must be an http:// or https:// URL
- rule: {"required":true}

### spec.notification.webhooks[].properties

`map<string, string>`

Extra properties merged into the webhook payload (e.g. a routing
key your receiver dispatches on). Azure's read API omits a webhook's
empty properties map -- provide at least one property when a
round-trip-stable configuration matters.

### spec.tags

`map<string, string>`

Tags to apply to the autoscale setting, merged over the
Planton-derived metadata tags (user values win on key conflicts).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMonitorAutoscaleSetting, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.autoscale_setting_id` | `string` | The autoscale setting's ARM resource ID (.../providers/Microsoft.Insights/autoScaleSettings/{name}). |
| `status.outputs.autoscale_setting_name` | `string` | The autoscale setting's resource name within its resource group. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## See Also

- [Overview](../README.md)
