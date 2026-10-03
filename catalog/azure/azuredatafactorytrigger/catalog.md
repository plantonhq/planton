# Azure Data Factory Trigger

Deploys one trigger inside an Azure Data Factory -- the instruction that starts pipelines automatically, in any of Azure's four types: schedule (wall-clock recurrence), tumbling window (once per contiguous time window, the backfill-friendly type), blob event (blobs created or deleted in a storage account), or custom event (events published to an Event Grid custom topic). All four types share one lifecycle: a trigger is either started (evaluating its condition and firing real, billed pipeline runs) or stopped, driven by the `activated` field.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions exactly one trigger of the type the spec's variant block declares:

- **Schedule trigger** -- fires on a wall-clock recurrence, optionally narrowed to specific minutes, hours, days, or monthly weekday occurrences
- **Tumbling window trigger** -- fires once per contiguous, non-overlapping time window from a fixed start; backfills past windows, retries, rate-limits, and can depend on other windows (or its own earlier ones)
- **Blob event trigger** -- fires when blobs are created or deleted in a storage account, filtered by path (Azure wires an Event Grid subscription on the storage account behind the scenes)
- **Custom event trigger** -- fires on events published to an Event Grid custom topic, filtered by subject and event type

Because Azure forbids editing a started trigger, both IaC engines stop the trigger before every update and delete, then start it again when `activated` is true.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Data Factory and Pipeline** -- the trigger lives in a factory and starts pipelines by name; reference an AzureDataFactory's `data_factory_id` output and AzureDataFactoryPipeline `pipeline_name` outputs.
- **Blob event**: an AzureStorageAccount reference. **Custom event**: an AzureEventgridTopic reference.

### Azure Subscription

- **Blob event triggers require the Microsoft.EventGrid resource provider registered** -- Azure creates the Event Grid subscription on the storage account behind the scenes.
- **Custom event triggers require an Event-Grid-schema topic** -- a CloudEvents-schema topic validates webhook subscribers with an HTTP OPTIONS handshake Data Factory's endpoint does not answer, so Start fails with "Webhook endpoint validation failed ... MethodNotAllowed".
- **A started trigger RUNS PIPELINES** -- deploying with `activated: true` (the default) and a due schedule starts real, billed pipeline runs immediately. Set the start time deliberately.

## Deploy

### Console

Open the deployment store, find **Azure Data Factory Trigger**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Daily Schedule**, **Hourly Tumbling Window**, **Blob Landing Event**, or **Custom Event** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactoryTrigger
metadata:
  name: nightly-ingest
  org: acme-corp
  env: prod
spec:
  dataFactoryId:
    valueFrom:
      kind: AzureDataFactory
      name: data-platform
      fieldPath: status.outputs.data_factory_id
  name: nightly-ingest
  activated: false
  schedule:
    frequency: Day
    interval: 1
    startTime: "2026-09-01T00:00:00Z"
    recurrenceSchedule:
      hours: [2]
      minutes: [0]
    pipelines:
      - name:
          valueFrom:
            kind: AzureDataFactoryPipeline
            name: ingest-daily
            fieldPath: status.outputs.pipeline_name
        parameters:
          windowStart: "@trigger().scheduledTime"
```

```shell
planton apply -f data-factory-trigger.yaml
```

This creates a daily 02:00 UTC schedule trigger targeting the `ingest-daily` pipeline, deployed STOPPED -- nothing runs until you flip `activated` to true. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying the factory, its pipelines, and their triggers as one chart, ValueFromRef wires the trigger to resources deployed in the same InfraPipeline:

```yaml
spec:
  dataFactoryId:
    valueFrom:
      kind: AzureDataFactory
      name: data-platform
      fieldPath: status.outputs.data_factory_id
  blobEvent:
    storageAccountId:
      valueFrom:
        kind: AzureStorageAccount
        name: landing
        fieldPath: status.outputs.storage_account_id
    events:
      - Microsoft.Storage.BlobCreated
    blobPathBeginsWith: /landing/blobs/raw/
    pipelines:
      - name:
          valueFrom:
            kind: AzureDataFactoryPipeline
            name: parse-landing
            fieldPath: status.outputs.pipeline_name
```

The InfraPipeline resolves the dependency graph -- factory, storage account, and pipeline first, then this trigger.

## Key Configuration

These are the most important decisions when configuring a trigger. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**A started trigger is a running system** -- `activated: true` (the default) means the trigger starts evaluating the moment it deploys, and a schedule trigger with an omitted `startTime` starts from NOW. Deploy new triggers with an explicit future start time or `activated: false`, verify the target pipeline's Debug run, then let it go live. One live-proven bound: ARM refuses to START a schedule trigger whose next execution is more than 18 months out -- park distant triggers with `activated: false`, not with a distant date.

**Schedule for clocks, tumbling window for data** -- a schedule trigger says "run at 02:00"; a tumbling window trigger says "process 00:00-01:00, then 01:00-02:00, ..." and passes each window's bounds to the pipeline (`@trigger().outputs.windowStartTime`). If the pipeline's work is a function of a time range -- every incremental load -- use tumbling windows: backfill by pointing `startTime` at history, per-window retry, `maxConcurrency` rate-limiting, and window dependencies. The tumbling window's `frequency`, `interval`, and `startTime` are fixed at creation -- changing them replaces the trigger.

**Backfills are deliberate, bounded operations** -- a tumbling window `startTime` in the past runs EVERY window between then and now. Before pointing at history: set `maxConcurrency` to what the sink tables tolerate (the default of 50 fans out fifty parallel runs), set the pipeline's own concurrency guard, and give the trigger `retry` so transient failures do not strand single windows in a sea of green ones. A self-dependency (a dependency entry with no trigger name and a negative `offset`) serializes windows for loads that must apply strictly in order.

**Blob event triggers have a second resource behind them** -- the Event Grid subscription on the storage account is part of your trigger's failure domain, and the path filters (`blobPathBeginsWith` / `blobPathEndsWith` -- at least one is required) are your only volume guard: an over-broad filter on a busy account fires the pipeline for every blob. Filter to the narrowest path that works, and set `ignoreEmptyBlobs` when upstream tools write zero-byte markers.

**Updates bounce the trigger -- plan for the gap** -- every update stops the trigger, applies the change, and starts it again. Schedule firings due DURING that gap are skipped, not queued: for a once-a-day trigger updated at the wrong moment, that is a missed day. Update triggers outside their firing windows, or reconcile the skipped window manually.

**A trigger holds its pipelines hostage at delete time** -- ARM refuses to delete a pipeline any trigger still references, including triggers in a failed half-started state. Delete triggers before their pipelines; deleting the whole factory removes everything inside it in one motion and sidesteps the ordering.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureDataFactory** | `dataFactoryId` | `status.outputs.data_factory_id` |
| **AzureDataFactoryPipeline** (every type) | `schedule.pipelines[].name`, `tumblingWindow.pipeline.name`, `blobEvent.pipelines[].name`, `customEvent.pipelines[].name` | `status.outputs.pipeline_name` |
| **AzureStorageAccount** (blob event) | `blobEvent.storageAccountId` | `status.outputs.storage_account_id` |
| **AzureEventgridTopic** (custom event) | `customEvent.eventgridTopicId` | `status.outputs.topic_id` |
| **AzureDataFactoryTrigger** (tumbling window dependencies) | `tumblingWindow.dependencies[].triggerName` | `status.outputs.trigger_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `trigger_name` | The trigger's name inside the factory | Other tumbling window triggers' `dependencies[].triggerName` -- chaining windows across triggers |
| `trigger_id` | The ARM ID (`{factory_id}/triggers/{name}`) | ARM-level references and import tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Nightly schedule, deployed stopped** -- daily at 02:00 UTC with `activated: false`, flipped on only after the pipeline's Debug run passes. Start from the **Daily Schedule** preset.

**Hourly incremental load** -- one contiguous hour per run, each run receiving its window's bounds through parameters, with retry and a deliberate concurrency cap. Start from the **Hourly Tumbling Window** preset.

**File-arrival trigger** -- fire the parse pipeline the moment a `.csv` lands under the landing path, passing the blob's folder and file name to the run. Start from the **Blob Landing Event** preset.

**Cross-system readiness event** -- fire the load pipeline when an upstream system publishes a readiness event to an Event Grid custom topic (Event Grid schema only). Start from the **Custom Event** preset.

## Works With

- [**Azure Data Factory**](/infra-catalog/azure-data-factory) -- the factory the trigger lives in, referenced by `dataFactoryId`
- [**Azure Data Factory Pipeline**](/infra-catalog/azure-data-factory-pipeline) -- what every trigger fires, referenced by pipeline name
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the event source for blob event triggers
- [**Azure Event Grid Topic**](/infra-catalog/azure-eventgrid-topic) -- the event source for custom event triggers (Event Grid schema required)
