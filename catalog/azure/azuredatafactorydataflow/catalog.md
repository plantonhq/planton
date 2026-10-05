# Azure Data Factory Data Flow

Deploys one data flow inside an Azure Data Factory -- a visually-designed transformation (filter, join, aggregate, reshape) that executes on Data Factory's managed Spark runtime when a pipeline runs it. One kind covers both provider forms, which share one schema and one name namespace inside the factory: a mapping data flow (the default) is the complete, runnable transformation with at least one source and one sink; a flowlet (`flowlet: true`) is the reusable snippet other data flows embed by reference, whose sources and sinks are optional because the embedding flow supplies them.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions exactly one of:

- **Mapping data flow** (`flowlet: false`, the default) -- the runnable transformation: the data flow script, named sources and sinks bound to datasets or linked services, and named intermediate transformations
- **Flowlet** (`flowlet: true`) -- the reusable snippet other data flows embed at their source, sink, or transformation points

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Data Factory** -- referenced through `dataFactoryId` as a literal ARM ID or an AzureDataFactory ValueFromRef; the data flow lives inside it.
- **The datasets and linked services the flow's endpoints bind to** must already exist in the factory -- they are validated when the flow is saved or run, not by the catalog.
- **A Studio-authored script** -- build the flow visually in the Data Factory Studio, open the "Script" view, and ship that artifact in `script` or `scriptLines`. Invalid scripts fail at deploy time.

## Deploy

### Console

Open the deployment store, find **Azure Data Factory Data Flow**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Mapping Transformation** or **Reusable Flowlet** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactoryDataFlow
metadata:
  name: transform-orders
  org: acme-corp
  env: prod
spec:
  dataFactoryId:
    value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.DataFactory/factories/acme-data-platform"
  name: transform-orders
  folder: transformations
  sources:
    - name: rawOrders
      linkedService:
        name:
          value: "raw-datalake"
  sinks:
    - name: curatedOrders
      linkedService:
        name:
          value: "curated-datalake"
  script: |
    source(
      allowSchemaDrift: true,
      validateSchema: false) ~> rawOrders
    rawOrders sink(
      allowSchemaDrift: true,
      validateSchema: false) ~> curatedOrders
```

```shell
planton apply -f data-factory-data-flow.yaml
```

This creates a mapping data flow with one source and one sink bound to the factory's linked services, carrying a pass-through script to replace with your Studio-authored flow. An Infra Job tracks the provisioning in real time.

### InfraChart

When a chart provisions the factory and its flows together, wire the factory -- and any embedded flowlet -- by reference:

```yaml
spec:
  dataFactoryId:
    valueFrom:
      kind: AzureDataFactory
      name: data-platform
      fieldPath: status.outputs.data_factory_id
  name: transform-orders
  sources:
    - name: rawOrders
      flowlet:
        name:
          valueFrom:
            kind: AzureDataFactoryDataFlow
            name: scrub-pii
            fieldPath: status.outputs.data_flow_name
  sinks:
    - name: curatedOrders
      linkedService:
        name:
          value: "curated-datalake"
  script: |
    source(
      allowSchemaDrift: true,
      validateSchema: false) ~> rawOrders
    rawOrders sink(
      allowSchemaDrift: true,
      validateSchema: false) ~> curatedOrders
```

The InfraPipeline resolves the dependency graph, deploys the factory and the flowlet first, then creates this flow with the resolved references.

## Key Configuration

These are the most important decisions when configuring a data flow. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Author in the Studio, ship through the catalog** -- The data flow script is not a language to hand-write: build the flow visually in the Data Factory Studio (a dev factory is ideal), open the "Script" view, and copy it into `script` -- or `scriptLines`, which diffs better under version control. The catalog deliberately does not re-model the script language; Azure owns it and evolves it on its own schedule. The `sources`/`sinks`/`transformations` blocks must name the same streams the script names -- a mismatch fails at deploy time.

**Deploy-time green is not run-time green** -- A data flow deploys the moment its script parses and its stream names line up, but the datasets and linked services its endpoints bind to are validated when a pipeline actually RUNS the flow. After deploying, run the owning pipeline's Debug execution once before trusting the flow in schedules.

**Mapping flow or flowlet is an identity decision** -- The two forms share one name namespace in the factory, and flipping `flowlet` replaces the object. Flowlets may omit sources and sinks; mapping flows require at least one of each (validated at manifest time). Data flows never run standalone -- a pipeline's Execute Data Flow activity runs them.

**Flowlets are your dedup lever -- name them like an API** -- A flowlet's name is its contract: every embedding flow references it by name, and renaming it breaks them all at their next deploy (the reference resolves at save time, not tracked by ARM). Treat flowlet names like package names -- stable, versioned by suffix when logic changes shape (`scrub-pii-v2`), never recycled for different logic.

**Spark spins up per run -- batch accordingly** -- Every pipeline run that executes a data flow pays a Spark cluster spin-up, minutes billed per vCore-hour. Ten small flows run separately cost ten spin-ups; the same logic as one flow with ten transformations costs one. Consolidate where the logic allows -- and note the compute size and TTL dials live on the pipeline's Execute Data Flow activity, not on this resource.

**Rejected-row routing beats failing the run** -- For flows ingesting third-party data, wire each sink's `rejectedLinkedService` to a quarantine store: schema-violating rows divert instead of failing the whole window's run. Rejected-data routing exists on sinks only -- Azure's model carries it nowhere else.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureDataFactory** | `dataFactoryId` | `status.outputs.data_factory_id` |
| **AzureDataFactoryDataset** (optional, per endpoint) | `sources[].dataset.name`, `sinks[].dataset.name`, `transformations[].dataset.name` | `status.outputs.dataset_name` |
| **AzureDataFactoryLinkedService** (optional, per endpoint) | `sources[].linkedService.name`, `sinks[].linkedService.name` (and the schema/rejected variants) | `status.outputs.linked_service_name` |
| **AzureDataFactoryDataFlow** (optional, embedded flowlets) | `sources[].flowlet.name`, `sinks[].flowlet.name`, `transformations[].flowlet.name` | `status.outputs.data_flow_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `data_flow_name` | The data flow's name in the factory | Other data flows' flowlet references -- how a flowlet gets embedded |

The other output, `data_flow_id`, is the flow's ARM ID (`{factory_id}/dataflows/{name}`, the same shape for both forms) -- pipelines reference flows by name inside the factory, so nothing consumes the ID by reference.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Curate the landing zone** -- one mapping flow per subject area, cleaning raw landing data into curated tables, run by the pipeline that ingests the raw data. Stream endpoints bind to linked services (or datasets); the real script comes from the Studio. Start from the **Mapping Transformation** preset.

**Shared cleanup as a flowlet** -- organization-wide rules (PII scrubbing, column standardization) authored once as a flowlet and embedded by every ingest flow -- identical logic everywhere instead of copy-pasted script blocks that drift. Start from the **Reusable Flowlet** preset.

## Works With

- [**Azure Data Factory**](/infra-catalog/azure-data-factory) -- the workspace the data flow lives in, referenced by its `data_factory_id` output
- [**Azure Data Factory Dataset**](/infra-catalog/azure-data-factory-dataset) -- named data views the flow's endpoints bind to
- [**Azure Data Factory Linked Service**](/infra-catalog/azure-data-factory-linked-service) -- store connections for dataset-less endpoints, schema drift, and rejected-row quarantine
- [**Azure Data Factory Pipeline**](/infra-catalog/azure-data-factory-pipeline) -- runs the flow through its Execute Data Flow activity; data flows never run standalone
