# AzureDataFactoryTrigger

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureDataFactoryTriggerSpec** defines an Azure Data Factory
trigger -- the instruction that starts pipelines automatically
inside a factory (AzureDataFactory).

The trigger type is declared by which variant block is present: set
exactly one of `schedule`, `tumbling_window`, `blob_event`, or
`custom_event` -- mirroring Azure's four trigger types.
  - `schedule`: fire on a wall-clock recurrence (every N minutes/
    hours/days/weeks/months, optionally narrowed to specific days
    and times).
  - `tumbling_window`: fire once per contiguous, non-overlapping
    time window from a fixed start time -- the backfill-friendly
    type that retries, rate-limits, and can depend on other
    tumbling triggers (or itself).
  - `blob_event`: fire when blobs are created or deleted in a
    storage account, filtered by path.
  - `custom_event`: fire on events published to an Event Grid
    custom topic, filtered by subject and event type.

All four types share one name namespace, and one LIFECYCLE: a
trigger is either started (evaluating its condition) or stopped.
`activated` drives that state -- and because Azure forbids editing
a started trigger, both engines stop the trigger before every
update and delete, then start it again when `activated` is true.

## Example

```yaml
# Deep-shape example for docs and offline validation: a schedule
# trigger exercising the full recurrence surface -- weekly frequency
# narrowed by the recurrence schedule (days, hours, minutes, monthly
# occurrences), start/end times, a time zone, and two pipelines with
# parameters. References are literal values so the manifest validates
# standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactoryTrigger
metadata:
  name: test-trigger
  id: test-trigger
  org: test-org
  env: test
spec:
  dataFactoryId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.DataFactory/factories/test-df
  name: weekly-loads
  description: Fires the weekly loads Monday and Friday at 02:00 and 14:30.
  annotations:
    - team:data
  activated: false
  schedule:
    frequency: Week
    interval: 1
    startTime: "2026-09-01T00:00:00Z"
    endTime: "2027-09-01T00:00:00Z"
    timeZone: UTC
    recurrenceSchedule:
      daysOfWeek:
        - Monday
        - Friday
      hours: [2, 14]
      minutes: [0, 30]
    pipelines:
      - name:
          value: ingest-weekly
        parameters:
          window: "@trigger().scheduledTime"
      - name:
          value: publish-weekly
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.dataFactoryId` | `string \| valueFrom` | yes |  | AzureDataFactory (`status.outputs.data_factory_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.annotations` | `[]string` |  |  |  |
| `spec.activated` | `bool` |  | `true` |  |
| `spec.schedule` | `AzureDataFactoryTriggerSchedule` |  |  |  |
| `spec.schedule.frequency` | `string` |  | `Minute` |  |
| `spec.schedule.interval` | `int32` |  | `1` |  |
| `spec.schedule.startTime` | `string` |  |  |  |
| `spec.schedule.endTime` | `string` |  |  |  |
| `spec.schedule.timeZone` | `string` |  |  |  |
| `spec.schedule.recurrenceSchedule` | `AzureDataFactoryTriggerRecurrenceSchedule` |  |  |  |
| `spec.schedule.recurrenceSchedule.daysOfMonth` | `[]int32` |  |  |  |
| `spec.schedule.recurrenceSchedule.daysOfWeek` | `[]string` |  |  |  |
| `spec.schedule.recurrenceSchedule.hours` | `[]int32` |  |  |  |
| `spec.schedule.recurrenceSchedule.minutes` | `[]int32` |  |  |  |
| `spec.schedule.recurrenceSchedule.monthly` | `[]AzureDataFactoryTriggerMonthlyOccurrence` |  |  |  |
| `spec.schedule.recurrenceSchedule.monthly[].weekday` | `string` | yes |  |  |
| `spec.schedule.recurrenceSchedule.monthly[].week` | `int32` |  |  |  |
| `spec.schedule.pipelines` | `[]AzureDataFactoryTriggerPipelineReference` | yes |  |  |
| `spec.schedule.pipelines[].name` | `string \| valueFrom` | yes |  | AzureDataFactoryPipeline (`status.outputs.pipeline_name`) |
| `spec.schedule.pipelines[].parameters` | `map<string, string>` |  |  |  |
| `spec.tumblingWindow` | `AzureDataFactoryTriggerTumblingWindow` |  |  |  |
| `spec.tumblingWindow.frequency` | `string` | yes |  |  |
| `spec.tumblingWindow.interval` | `int32` |  |  |  |
| `spec.tumblingWindow.startTime` | `string` | yes |  |  |
| `spec.tumblingWindow.endTime` | `string` |  |  |  |
| `spec.tumblingWindow.delay` | `string` |  |  |  |
| `spec.tumblingWindow.maxConcurrency` | `int32` |  | `50` |  |
| `spec.tumblingWindow.retry` | `AzureDataFactoryTriggerRetryPolicy` |  |  |  |
| `spec.tumblingWindow.retry.count` | `int32` |  |  |  |
| `spec.tumblingWindow.retry.interval` | `int32` |  | `30` |  |
| `spec.tumblingWindow.dependencies` | `[]AzureDataFactoryTriggerDependency` |  |  |  |
| `spec.tumblingWindow.dependencies[].triggerName` | `string \| valueFrom` |  |  | AzureDataFactoryTrigger (`status.outputs.trigger_name`) |
| `spec.tumblingWindow.dependencies[].offset` | `string` |  |  |  |
| `spec.tumblingWindow.dependencies[].size` | `string` |  |  |  |
| `spec.tumblingWindow.additionalProperties` | `map<string, string>` |  |  |  |
| `spec.tumblingWindow.pipeline` | `AzureDataFactoryTriggerPipelineReference` | yes |  |  |
| `spec.tumblingWindow.pipeline.name` | `string \| valueFrom` | yes |  | AzureDataFactoryPipeline (`status.outputs.pipeline_name`) |
| `spec.tumblingWindow.pipeline.parameters` | `map<string, string>` |  |  |  |
| `spec.blobEvent` | `AzureDataFactoryTriggerBlobEvent` |  |  |  |
| `spec.blobEvent.storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.blobEvent.events` | `[]string` | yes |  |  |
| `spec.blobEvent.blobPathBeginsWith` | `string` |  |  |  |
| `spec.blobEvent.blobPathEndsWith` | `string` |  |  |  |
| `spec.blobEvent.ignoreEmptyBlobs` | `bool` |  | `false` |  |
| `spec.blobEvent.additionalProperties` | `map<string, string>` |  |  |  |
| `spec.blobEvent.pipelines` | `[]AzureDataFactoryTriggerPipelineReference` | yes |  |  |
| `spec.blobEvent.pipelines[].name` | `string \| valueFrom` | yes |  | AzureDataFactoryPipeline (`status.outputs.pipeline_name`) |
| `spec.blobEvent.pipelines[].parameters` | `map<string, string>` |  |  |  |
| `spec.customEvent` | `AzureDataFactoryTriggerCustomEvent` |  |  |  |
| `spec.customEvent.eventgridTopicId` | `string \| valueFrom` | yes |  | AzureEventgridTopic (`status.outputs.topic_id`) |
| `spec.customEvent.events` | `[]string` | yes |  |  |
| `spec.customEvent.subjectBeginsWith` | `string` |  |  |  |
| `spec.customEvent.subjectEndsWith` | `string` |  |  |  |
| `spec.customEvent.additionalProperties` | `map<string, string>` |  |  |  |
| `spec.customEvent.pipelines` | `[]AzureDataFactoryTriggerPipelineReference` | yes |  |  |
| `spec.customEvent.pipelines[].name` | `string \| valueFrom` | yes |  | AzureDataFactoryPipeline (`status.outputs.pipeline_name`) |
| `spec.customEvent.pipelines[].parameters` | `map<string, string>` |  |  |  |

## Field Details

### spec.dataFactoryId

`string | valueFrom` · required

The Data Factory the trigger lives in, by ARM ID. Can be a
literal string or a reference to an AzureDataFactory output.

**ForceNew**: changing this destroys and recreates the trigger.

- references: AzureDataFactory (`status.outputs.data_factory_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactory, name: <that resource's name>, fieldPath: status.outputs.data_factory_id}} -- a bare string does not parse

### spec.name

`string` · required

The trigger's name -- unique within the factory across all four
types. Must start with a letter, number, or underscore, and must
not contain any of < > * # . % & : \ + ? / (Azure's trigger
naming rules, the same class as pipeline names).

**ForceNew**: changing this destroys and recreates the trigger.

- rule: Trigger names must start with a letter, number, or underscore and must not contain < > * # . % & : \ + ? /
- rule: {"required":true}

### spec.description

`string`

A human-readable description of what the trigger does.

### spec.annotations

`[]string`

Free-form annotation strings stored on the trigger.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.activated

`bool` · optional (explicit presence)

Whether the trigger is STARTED (evaluating its condition and
firing pipeline runs). Unspecified applies true, the provider's
default -- set false to deploy the trigger stopped. Updatable in
place: flipping it starts or stops the trigger via Azure's
Start/Stop API (the engines stop a started trigger before any
update, then re-start it).

- default: `true`

### spec.schedule

`AzureDataFactoryTriggerSchedule`

Fire on a wall-clock recurrence. Set exactly one variant block on
this spec.

### spec.schedule.frequency

`string` · optional (explicit presence)

The recurrence unit. Unspecified applies "Minute", the provider's
default -- the modules always send the effective value
explicitly.

- default: `Minute`
- rule: {"string":{"in":["Minute","Hour","Day","Week","Month"]}}

### spec.schedule.interval

`int32` · optional (explicit presence)

How many `frequency` units between firings (e.g. frequency
"Hour" + interval 6 fires every six hours). Unspecified applies
1 -- the modules always send the effective value explicitly.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.schedule.startTime

`string`

When the recurrence begins, as an RFC 3339 timestamp in UTC
(e.g. "2026-09-01T00:00:00Z"). Omit to start from the moment of
deployment (Azure fills in the current time). A time in the past
starts the recurrence immediately.

LIVE CONTRACT (enforced by ARM when the trigger STARTS, not at
save): the recurrence's NEXT execution must fall within 18
months of the current time, or Start fails with
InvalidWorkflowTriggerRecurrenceSchedule. A far-future
start_time therefore cannot be combined with `activated: true`;
a past start_time is always startable (the next execution is
computed forward from the recurrence).

- rule: start_time must be an RFC 3339 timestamp, e.g. 2026-09-01T00:00:00Z

### spec.schedule.endTime

`string`

When the recurrence ends, as an RFC 3339 timestamp in UTC. Omit
for no end.

- rule: end_time must be an RFC 3339 timestamp, e.g. 2026-12-31T00:00:00Z

### spec.schedule.timeZone

`string`

The time zone the recurrence is evaluated in (e.g. "UTC",
"Eastern Standard Time" -- Windows time-zone IDs). Omit for UTC.

### spec.schedule.recurrenceSchedule

`AzureDataFactoryTriggerRecurrenceSchedule`

Narrows the recurrence to specific minutes, hours, days of the
week/month, or monthly weekday occurrences -- meaningful for
Week and Month frequencies (Azure evaluates it there). The
provider's `schedule` block.

### spec.schedule.recurrenceSchedule.daysOfMonth

`[]int32`

Days of the month to fire on: 1-31 counted from the month's
start, or -1..-31 counted from its end (-1 is the last day).

- rule: {"repeated":{"items":{"cel":[{"id":"data_factory_trigger_days_of_month_range","message":"days_of_month entries must be 1-31 (from the start) or -1..-31 (from the end)","expression":"this != 0 && this >= -31 && this <= 31"}]}}}

### spec.schedule.recurrenceSchedule.daysOfWeek

`[]string`

Days of the week to fire on, e.g. "Monday" (at most all seven).

- rule: {"repeated":{"maxItems":"7","items":{"string":{"in":["Monday","Tuesday","Wednesday","Thursday","Friday","Saturday","Sunday"]}}}}

### spec.schedule.recurrenceSchedule.hours

`[]int32`

Hours of the day to fire at (0-24 -- the provider's own accepted
range, mirrored exactly).

- rule: {"repeated":{"items":{"int32":{"lte":24,"gte":0}}}}

### spec.schedule.recurrenceSchedule.minutes

`[]int32`

Minutes of the hour to fire at (0-60 -- the provider's own
accepted range, mirrored exactly).

- rule: {"repeated":{"items":{"int32":{"lte":60,"gte":0}}}}

### spec.schedule.recurrenceSchedule.monthly

`[]AzureDataFactoryTriggerMonthlyOccurrence`

Monthly weekday occurrences to fire on (e.g. the second Friday:
weekday "Friday", week 2) -- for Month frequency.

### spec.schedule.recurrenceSchedule.monthly[].weekday

`string` · required

The weekday, e.g. "Friday".

- rule: {"required":true,"string":{"in":["Monday","Tuesday","Wednesday","Thursday","Friday","Saturday","Sunday"]}}

### spec.schedule.recurrenceSchedule.monthly[].week

`int32` · optional (explicit presence)

Which occurrence in the month: 1-5 counted from the month's
start, or -1..-5 counted from its end (-1 is the last). Omit for
every occurrence of the weekday.

- rule: week must be 1-5 (from the start) or -1..-5 (from the end)

### spec.schedule.pipelines

`[]AzureDataFactoryTriggerPipelineReference` · required

The pipelines this trigger starts on every firing, each with its
parameter values. At least one -- a trigger with nothing to start
is rejected by the provider.

- rule: {"repeated":{"minItems":"1"}}

### spec.schedule.pipelines[].name

`string | valueFrom` · required

The pipeline's name inside the factory -- defaults to referencing
an AzureDataFactoryPipeline's pipeline_name output.

- references: AzureDataFactoryPipeline (`status.outputs.pipeline_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryPipeline, name: <that resource's name>, fieldPath: status.outputs.pipeline_name}} -- a bare string does not parse

### spec.schedule.pipelines[].parameters

`map<string, string>`

Values for the pipeline's parameters, keyed by parameter name.
Trigger system variables (e.g. @trigger().outputs.windowStartTime
for tumbling windows, @triggerBody().fileName for blob events)
are valid values.

### spec.tumblingWindow

`AzureDataFactoryTriggerTumblingWindow`

Fire once per contiguous time window from a fixed start. Set
exactly one variant block on this spec.

### spec.tumblingWindow.frequency

`string` · required

The window unit. FIXED AT CREATION -- changing it replaces the
trigger.

- rule: {"required":true,"string":{"in":["Minute","Hour","Month"]}}

### spec.tumblingWindow.interval

`int32`

The window length in `frequency` units. FIXED AT CREATION --
changing it replaces the trigger.

- rule: {"int32":{"gte":1}}

### spec.tumblingWindow.startTime

`string` · required

The first window's start, as an RFC 3339 timestamp in UTC --
REQUIRED (windows are counted from here; a past start backfills).
FIXED AT CREATION -- changing it replaces the trigger.

- rule: start_time must be an RFC 3339 timestamp, e.g. 2026-09-01T00:00:00Z
- rule: {"required":true}

### spec.tumblingWindow.endTime

`string`

The last window's end, as an RFC 3339 timestamp in UTC. Omit for
no end.

- rule: end_time must be an RFC 3339 timestamp, e.g. 2026-12-31T00:00:00Z

### spec.tumblingWindow.delay

`string`

How long past each window's end the firing waits (for late data),
as a Data Factory TimeSpan: "hh:mm:ss", optionally with a day
prefix ("1.06:00:00") or a leading minus. Omit for no delay.

- rule: delay must be a TimeSpan like 00:10:00 or 1.06:00:00

### spec.tumblingWindow.maxConcurrency

`int32` · optional (explicit presence)

How many windows may run at once, 1-50 (backfills fan out up to
this). Unspecified applies 50, the provider's default -- the
modules always send the effective value explicitly.

- default: `50`
- rule: {"int32":{"lte":50,"gte":1}}

### spec.tumblingWindow.retry

`AzureDataFactoryTriggerRetryPolicy`

Retry policy for failed window runs. Omit for no retries.

### spec.tumblingWindow.retry.count

`int32`

The maximum number of retry attempts per window (at least 1).

- rule: {"int32":{"gte":1}}

### spec.tumblingWindow.retry.interval

`int32` · optional (explicit presence)

Seconds between attempts. Unspecified applies 30, the provider's
default -- the modules always send the effective value
explicitly.

- default: `30`
- rule: {"int32":{"gte":0}}

### spec.tumblingWindow.dependencies

`[]AzureDataFactoryTriggerDependency`

Windows this trigger waits on before firing: other tumbling
triggers' windows, or this trigger's own earlier windows (a
SELF-dependency -- omit trigger_name for that).

### spec.tumblingWindow.dependencies[].triggerName

`string | valueFrom`

The tumbling window trigger depended on, by name -- defaults to
referencing an AzureDataFactoryTrigger's trigger_name output.
OMIT for a self-dependency (this trigger's own earlier windows).

- references: AzureDataFactoryTrigger (`status.outputs.trigger_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryTrigger, name: <that resource's name>, fieldPath: status.outputs.trigger_name}} -- a bare string does not parse

### spec.tumblingWindow.dependencies[].offset

`string`

Shifts which of the dependency's windows is waited on, as a
TimeSpan (e.g. "-24:00:00" waits on yesterday's window). REQUIRED
by Azure for self-dependencies (it must be negative there). Omit
for the aligned window.

- rule: offset must be a TimeSpan like 24:00:00 or -24:00:00

### spec.tumblingWindow.dependencies[].size

`string`

The size of the dependency window waited on, as a TimeSpan. Omit
to match the depended-on trigger's own window size.

- rule: size must be a TimeSpan like 06:00:00

### spec.tumblingWindow.additionalProperties

`map<string, string>`

Additional type properties passed through to Azure as-is --
Data Factory's escape hatch for window properties the schema does
not model.

### spec.tumblingWindow.pipeline

`AzureDataFactoryTriggerPipelineReference` · required

The ONE pipeline this trigger runs per window (tumbling window
triggers drive exactly one pipeline -- Azure's own model), with
its parameter values. Window-scoped system variables
(trigger().outputs.windowStartTime / windowEndTime) are passed
through parameters.

- rule: {"required":true}

### spec.tumblingWindow.pipeline.name

`string | valueFrom` · required

The pipeline's name inside the factory -- defaults to referencing
an AzureDataFactoryPipeline's pipeline_name output.

- references: AzureDataFactoryPipeline (`status.outputs.pipeline_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryPipeline, name: <that resource's name>, fieldPath: status.outputs.pipeline_name}} -- a bare string does not parse

### spec.tumblingWindow.pipeline.parameters

`map<string, string>`

Values for the pipeline's parameters, keyed by parameter name.
Trigger system variables (e.g. @trigger().outputs.windowStartTime
for tumbling windows, @triggerBody().fileName for blob events)
are valid values.

### spec.blobEvent

`AzureDataFactoryTriggerBlobEvent`

Fire on blob creation/deletion in a storage account. Set exactly
one variant block on this spec.

- rule: Set blob_path_begins_with, blob_path_ends_with, or both

### spec.blobEvent.storageAccountId

`string | valueFrom` · required

The storage account whose blob events fire this trigger, by ARM
ID -- defaults to referencing an AzureStorageAccount's
storage_account_id output. FIXED AT CREATION -- changing it
replaces the trigger.

The trigger LISTENS to this account's blob events and belongs to its
factory, so on a diagram the reference is access, not placement.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.blobEvent.events

`[]string` · required

Which blob events fire the trigger (at least one).

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["Microsoft.Storage.BlobCreated","Microsoft.Storage.BlobDeleted"]}}}}

### spec.blobEvent.blobPathBeginsWith

`string`

Fire only for blob paths starting with this prefix (e.g.
"/container/blobs/raw/"). Set this, blob_path_ends_with, or both.

### spec.blobEvent.blobPathEndsWith

`string`

Fire only for blob paths ending with this suffix (e.g. ".csv").
Set this, blob_path_begins_with, or both.

### spec.blobEvent.ignoreEmptyBlobs

`bool` · optional (explicit presence)

Whether zero-byte blobs are ignored (skips the placeholder blobs
some tools write). Unspecified applies false -- the modules
always send the effective value explicitly.

- default: `false`

### spec.blobEvent.additionalProperties

`map<string, string>`

Additional type properties passed through to Azure as-is.

### spec.blobEvent.pipelines

`[]AzureDataFactoryTriggerPipelineReference` · required

The pipelines this trigger starts on every firing, each with its
parameter values. At least one. Event metadata
(@triggerBody().folderPath / fileName) is passed through
parameters.

- rule: {"repeated":{"minItems":"1"}}

### spec.blobEvent.pipelines[].name

`string | valueFrom` · required

The pipeline's name inside the factory -- defaults to referencing
an AzureDataFactoryPipeline's pipeline_name output.

- references: AzureDataFactoryPipeline (`status.outputs.pipeline_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryPipeline, name: <that resource's name>, fieldPath: status.outputs.pipeline_name}} -- a bare string does not parse

### spec.blobEvent.pipelines[].parameters

`map<string, string>`

Values for the pipeline's parameters, keyed by parameter name.
Trigger system variables (e.g. @trigger().outputs.windowStartTime
for tumbling windows, @triggerBody().fileName for blob events)
are valid values.

### spec.customEvent

`AzureDataFactoryTriggerCustomEvent`

Fire on events published to an Event Grid custom topic. Set
exactly one variant block on this spec.

### spec.customEvent.eventgridTopicId

`string | valueFrom` · required

The Event Grid custom topic whose events fire this trigger, by
ARM ID -- defaults to referencing an AzureEventgridTopic's
topic_id output. FIXED AT CREATION -- changing it replaces the
trigger.

LIVE CONTRACT: the topic must use the "EventGridSchema" input
schema. Data Factory subscribes to the topic with a webhook when
the trigger STARTS, and a CloudEvents-schema topic validates
webhook subscribers with an HTTP OPTIONS handshake Data
Factory's endpoint does not answer -- Start fails with "Webhook
endpoint validation failed ... MethodNotAllowed".

The trigger lives in its factory and listens TO the topic, so the
reference is access, not placement, on a diagram.

- references: AzureEventgridTopic (`status.outputs.topic_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventgridTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.customEvent.events

`[]string` · required

Which event types fire the trigger (at least one) -- free-form
strings matched against the published events' eventType field
(custom topics define their own vocabulary).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.customEvent.subjectBeginsWith

`string`

Fire only for events whose subject starts with this prefix. Omit
for no prefix filter.

### spec.customEvent.subjectEndsWith

`string`

Fire only for events whose subject ends with this suffix. Omit
for no suffix filter.

### spec.customEvent.additionalProperties

`map<string, string>`

Additional type properties passed through to Azure as-is.

### spec.customEvent.pipelines

`[]AzureDataFactoryTriggerPipelineReference` · required

The pipelines this trigger starts on every firing, each with its
parameter values. At least one. Event payload
(@triggerBody().event.data) is passed through parameters.

- rule: {"repeated":{"minItems":"1"}}

### spec.customEvent.pipelines[].name

`string | valueFrom` · required

The pipeline's name inside the factory -- defaults to referencing
an AzureDataFactoryPipeline's pipeline_name output.

- references: AzureDataFactoryPipeline (`status.outputs.pipeline_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryPipeline, name: <that resource's name>, fieldPath: status.outputs.pipeline_name}} -- a bare string does not parse

### spec.customEvent.pipelines[].parameters

`map<string, string>`

Values for the pipeline's parameters, keyed by parameter name.
Trigger system variables (e.g. @trigger().outputs.windowStartTime
for tumbling windows, @triggerBody().fileName for blob events)
are valid values.

## Validation Rules

- `azure_data_factory_trigger_exactly_one_variant`: Set exactly one trigger variant -- schedule, tumbling_window, blob_event, or custom_event -- the variant determines the trigger type

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDataFactoryTrigger, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.trigger_id` | `string` | The trigger's Azure Resource Manager ID ({factory_id}/triggers/{name}) -- the same ID shape for all four trigger types. |
| `status.outputs.trigger_name` | `string` | The trigger's name -- what other tumbling window triggers' dependency entries resolve against. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.dataFactoryId` | AzureDataFactory | `status.outputs.data_factory_id` |
| `spec.schedule.pipelines[].name` | AzureDataFactoryPipeline | `status.outputs.pipeline_name` |
| `spec.tumblingWindow.dependencies[].triggerName` | AzureDataFactoryTrigger | `status.outputs.trigger_name` |
| `spec.tumblingWindow.pipeline.name` | AzureDataFactoryPipeline | `status.outputs.pipeline_name` |
| `spec.blobEvent.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.blobEvent.pipelines[].name` | AzureDataFactoryPipeline | `status.outputs.pipeline_name` |
| `spec.customEvent.eventgridTopicId` | AzureEventgridTopic | `status.outputs.topic_id` |
| `spec.customEvent.pipelines[].name` | AzureDataFactoryPipeline | `status.outputs.pipeline_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureDataFactoryTrigger | `spec.tumblingWindow.dependencies[].triggerName` | `status.outputs.trigger_name` |

## See Also

- [Overview](../README.md)
