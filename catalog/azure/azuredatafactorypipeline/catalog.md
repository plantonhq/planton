# Azure Data Factory Pipeline

Deploys one pipeline inside an Azure Data Factory -- an ordered set of activities (copy data, run a data flow, call a stored procedure, wait, branch) that executes as a whole when triggered. The activities travel as raw JSON: exactly the "activities" array from the Data Factory Studio's Code view, because Azure owns that schema -- dozens of activity types, each with its own shape -- and the catalog deliberately does not re-model it. The Studio is the authoring surface; the manifest is the shipping surface.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Data Factory pipeline** -- the pipeline definition: activities JSON, run-time parameters with defaults, variables, annotations, the concurrency cap, the Studio display folder, and the elapsed-time metric threshold

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Data Factory** -- the pipeline lives in a factory; reference an AzureDataFactory's `data_factory_id` output or provide the ARM ID directly.

### Azure Subscription

- **Azure owns the activity schema** -- invalid JSON or unknown activity shapes fail at deploy time, not at validation time. Author activities in the Data Factory Studio and copy its Code view.
- **Linked services and datasets the activities reference must already exist** in the factory -- they are validated at RUN time, so a missing reference deploys green and fails on first trigger.
- **Parameters and variables are String-typed on this surface** -- declare other-typed parameters inside the activities JSON if you need them.

## Deploy

### Console

Open the deployment store, find **Azure Data Factory Pipeline**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Scheduled Copy** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactoryPipeline
metadata:
  name: ingest-daily
  org: acme-corp
  env: prod
spec:
  dataFactoryId:
    valueFrom:
      kind: AzureDataFactory
      name: data-platform
      fieldPath: status.outputs.data_factory_id
  name: ingest-daily
  folder: ingest
  concurrency: 1
  parameters:
    windowStart: ""
  activitiesJson: |
    [{"name": "placeholder-wait", "type": "Wait", "typeProperties": {"waitTimeInSeconds": 10}}]
```

```shell
planton apply -f data-factory-pipeline.yaml
```

This creates a single-concurrency pipeline named `ingest-daily` under the factory's `ingest` folder with one window parameter and a placeholder Wait activity -- swap in your Studio-authored activities array before wiring triggers. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying the factory and its pipelines as one chart, ValueFromRef wires the pipeline to the factory deployed in the same InfraPipeline:

```yaml
spec:
  dataFactoryId:
    valueFrom:
      kind: AzureDataFactory
      name: data-platform
      fieldPath: status.outputs.data_factory_id
```

The InfraPipeline resolves the dependency graph -- factory first, then this pipeline.

## Key Configuration

These are the most important decisions when configuring a pipeline. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Author in the Studio, ship through the catalog** -- `activitiesJson` is not a place to hand-write orchestration: build the pipeline in the Data Factory Studio (a dev factory is ideal), open the Code view, and copy the "activities" array into the manifest. Key ordering inside the JSON is not meaningful -- both IaC engines normalize it when diffing.

**Deploy-time green is not run-time green** -- a pipeline deploys successfully the moment its JSON parses and its activity types are known to ARM, but the linked services and datasets its activities reference are validated at run time. A pipeline can sit green for weeks and fail on first trigger because a referenced connection never existed. Run a Debug execution before wiring triggers.

**Concurrency is a self-protection dial, not a throughput dial** -- `concurrency` (1-50) caps simultaneous runs of THIS pipeline. Leave it unset and a backfill trigger can launch dozens of overlapping runs against the same tables; set it to 1 for pipelines that must never overlap (most incremental loads). Queued runs wait -- they do not fail.

**Parameterize windows, not environments** -- run-time `parameters` are for what changes per RUN (the date window, the source partition); what changes per ENVIRONMENT (connection strings, account names) belongs in the factory's global parameters and linked services. A pipeline whose JSON embeds environment facts cannot promote between factories, defeating the one-definition-many-environments model the factory/pipeline split exists for.

**Name and factory are one-way doors** -- `name` and `dataFactoryId` are ForceNew: changing either destroys and recreates the pipeline. Triggers reference the pipeline by name, so a rename orphans every trigger pointing at the old one.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureDataFactory** | `dataFactoryId` | `status.outputs.data_factory_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `pipeline_name` | The pipeline's name inside the factory | AzureDataFactoryTrigger's `pipelineName` -- every trigger names the pipeline it fires |
| `pipeline_id` | The ARM ID (`{factory_id}/pipelines/{name}`) | ARM-level references and pipeline-run API calls |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Window-parameterized incremental load** -- a single-concurrency pipeline with a `windowStart` parameter the trigger supplies per run; the standard incremental-ingestion shell. Start from the **Scheduled Copy** preset.

**Debug-first rollout** -- deploy the pipeline, run it once with Debug in the Studio to validate every linked service and dataset reference, then attach the trigger. The trigger is a separate resource, so the pipeline can bake as long as needed.

**One definition, many environments** -- keep environment facts out of the activities JSON and promote the same pipeline manifest across dev and prod factories, changing only the `dataFactoryId` reference.

## Works With

- [**Azure Data Factory**](/infra-catalog/azure-data-factory) -- the factory the pipeline lives in, referenced by `dataFactoryId`
- [**Azure Data Factory Trigger**](/infra-catalog/azure-data-factory-trigger) -- schedules, tumbling windows, and event triggers fire this pipeline by name
- [**Azure Data Factory Dataset**](/infra-catalog/azure-data-factory-dataset) -- the sources and sinks copy activities read and write
- [**Azure Data Factory Linked Service**](/infra-catalog/azure-data-factory-linked-service) -- the connections activities authenticate through
- [**Azure Data Factory Data Flow**](/infra-catalog/azure-data-factory-data-flow) -- execute-data-flow activities run transformations built as data flows
