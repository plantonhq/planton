# AzureDataFactoryDataFlow

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureDataFactoryDataFlowSpec** defines an Azure Data Factory data
flow -- a visually-designed data transformation (filter, join,
aggregate, reshape) that executes on Data Factory's managed Spark
runtime when a pipeline runs it.

One kind covers BOTH provider forms -- they share one schema and one
name namespace inside the factory:
  - a **mapping data flow** (the default): a complete
    transformation with at least one source and one sink;
  - a **flowlet** (`flowlet: true`): a reusable snippet other data
    flows embed by reference; its sources and sinks are optional
    because the embedding flow supplies them.

The transformation logic itself travels in `script` (or
`script_lines`) -- the data flow script the Data Factory Studio's
"Script" view shows. Azure owns that language; the catalog does not
re-model it. The `sources`/`sinks`/`transformations` blocks declare
the named endpoints the script references, each optionally bound to
a dataset, a linked service, or another flowlet.

## Example

```yaml
# Deep-shape example for docs and offline validation: a mapping data
# flow exercising the full endpoint surface -- dataset, linked service,
# schema linked service, and flowlet bindings on the source, rejected-
# row routing on the sink, an intermediate transformation, script
# lines, annotations, and the Studio folder. References are literal
# values so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactoryDataFlow
metadata:
  name: test-data-flow
  id: test-data-flow
  org: test-org
  env: test
spec:
  dataFactoryId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.DataFactory/factories/test-df
  name: transform-orders
  description: Cleans and conforms raw orders into the curated zone.
  folder: transformations/daily
  annotations:
    - team:data
    - tier:silver
  sources:
    - name: rawOrders
      description: Raw landing-zone orders.
      dataset:
        name:
          value: raw_orders
        parameters:
          path: raw/orders
      schemaLinkedService:
        name:
          value: schema-store
    - name: enrichment
      description: Shared cleanup embedded as a flowlet.
      flowlet:
        name:
          value: scrub-pii
        parameters:
          mode: strict
        datasetParameters: "path: 'raw/orders'"
  sinks:
    - name: curatedOrders
      description: The curated zone table.
      linkedService:
        name:
          value: lakehouse
        parameters:
          container: curated
      rejectedLinkedService:
        name:
          value: quarantine
  transformations:
    - name: joinReference
      description: Joins the reference data.
      linkedService:
        name:
          value: reference-store
  scriptLines:
    - "source(allowSchemaDrift: true, validateSchema: false) ~> rawOrders"
    - "source(allowSchemaDrift: true) ~> enrichment"
    - "rawOrders, enrichment join(orders.id == enrichment.id, joinType:'inner') ~> joinReference"
    - "joinReference sink(allowSchemaDrift: true, validateSchema: false) ~> curatedOrders"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.dataFactoryId` | `string \| valueFrom` | yes |  | AzureDataFactory (`status.outputs.data_factory_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.flowlet` | `bool` |  |  |  |
| `spec.script` | `string` |  |  |  |
| `spec.scriptLines` | `[]string` |  |  |  |
| `spec.sources` | `[]AzureDataFactoryDataFlowSource` |  |  |  |
| `spec.sources[].name` | `string` | yes |  |  |
| `spec.sources[].description` | `string` |  |  |  |
| `spec.sources[].dataset` | `AzureDataFactoryDataFlowDatasetReference` |  |  |  |
| `spec.sources[].dataset.name` | `string \| valueFrom` | yes |  | AzureDataFactoryDataset (`status.outputs.dataset_name`) |
| `spec.sources[].dataset.parameters` | `map<string, string>` |  |  |  |
| `spec.sources[].flowlet` | `AzureDataFactoryDataFlowFlowletReference` |  |  |  |
| `spec.sources[].flowlet.name` | `string \| valueFrom` | yes |  | AzureDataFactoryDataFlow (`status.outputs.data_flow_name`) |
| `spec.sources[].flowlet.parameters` | `map<string, string>` |  |  |  |
| `spec.sources[].flowlet.datasetParameters` | `string` |  |  |  |
| `spec.sources[].linkedService` | `AzureDataFactoryDataFlowLinkedServiceReference` |  |  |  |
| `spec.sources[].linkedService.name` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.sources[].linkedService.parameters` | `map<string, string>` |  |  |  |
| `spec.sources[].schemaLinkedService` | `AzureDataFactoryDataFlowLinkedServiceReference` |  |  |  |
| `spec.sources[].schemaLinkedService.name` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.sources[].schemaLinkedService.parameters` | `map<string, string>` |  |  |  |
| `spec.sinks` | `[]AzureDataFactoryDataFlowSink` |  |  |  |
| `spec.sinks[].name` | `string` | yes |  |  |
| `spec.sinks[].description` | `string` |  |  |  |
| `spec.sinks[].dataset` | `AzureDataFactoryDataFlowDatasetReference` |  |  |  |
| `spec.sinks[].dataset.name` | `string \| valueFrom` | yes |  | AzureDataFactoryDataset (`status.outputs.dataset_name`) |
| `spec.sinks[].dataset.parameters` | `map<string, string>` |  |  |  |
| `spec.sinks[].flowlet` | `AzureDataFactoryDataFlowFlowletReference` |  |  |  |
| `spec.sinks[].flowlet.name` | `string \| valueFrom` | yes |  | AzureDataFactoryDataFlow (`status.outputs.data_flow_name`) |
| `spec.sinks[].flowlet.parameters` | `map<string, string>` |  |  |  |
| `spec.sinks[].flowlet.datasetParameters` | `string` |  |  |  |
| `spec.sinks[].linkedService` | `AzureDataFactoryDataFlowLinkedServiceReference` |  |  |  |
| `spec.sinks[].linkedService.name` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.sinks[].linkedService.parameters` | `map<string, string>` |  |  |  |
| `spec.sinks[].schemaLinkedService` | `AzureDataFactoryDataFlowLinkedServiceReference` |  |  |  |
| `spec.sinks[].schemaLinkedService.name` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.sinks[].schemaLinkedService.parameters` | `map<string, string>` |  |  |  |
| `spec.sinks[].rejectedLinkedService` | `AzureDataFactoryDataFlowLinkedServiceReference` |  |  |  |
| `spec.sinks[].rejectedLinkedService.name` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.sinks[].rejectedLinkedService.parameters` | `map<string, string>` |  |  |  |
| `spec.transformations` | `[]AzureDataFactoryDataFlowTransformation` |  |  |  |
| `spec.transformations[].name` | `string` | yes |  |  |
| `spec.transformations[].description` | `string` |  |  |  |
| `spec.transformations[].dataset` | `AzureDataFactoryDataFlowDatasetReference` |  |  |  |
| `spec.transformations[].dataset.name` | `string \| valueFrom` | yes |  | AzureDataFactoryDataset (`status.outputs.dataset_name`) |
| `spec.transformations[].dataset.parameters` | `map<string, string>` |  |  |  |
| `spec.transformations[].flowlet` | `AzureDataFactoryDataFlowFlowletReference` |  |  |  |
| `spec.transformations[].flowlet.name` | `string \| valueFrom` | yes |  | AzureDataFactoryDataFlow (`status.outputs.data_flow_name`) |
| `spec.transformations[].flowlet.parameters` | `map<string, string>` |  |  |  |
| `spec.transformations[].flowlet.datasetParameters` | `string` |  |  |  |
| `spec.transformations[].linkedService` | `AzureDataFactoryDataFlowLinkedServiceReference` |  |  |  |
| `spec.transformations[].linkedService.name` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.transformations[].linkedService.parameters` | `map<string, string>` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.annotations` | `[]string` |  |  |  |
| `spec.folder` | `string` |  |  |  |

## Field Details

### spec.dataFactoryId

`string | valueFrom` · required

The Data Factory the data flow lives in, by ARM ID. Can be a
literal string or a reference to an AzureDataFactory output.

**ForceNew**: changing this destroys and recreates the data flow.

- references: AzureDataFactory (`status.outputs.data_factory_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactory, name: <that resource's name>, fieldPath: status.outputs.data_factory_id}} -- a bare string does not parse

### spec.name

`string` · required

The data flow's name -- unique within the factory across BOTH
forms (mapping data flows and flowlets share one namespace).
Azure's Data Factory naming rules apply at deploy time (must
start with a letter, number, or underscore; no
< > * # . % & : \ + ? / characters) -- the provider itself does
not pre-validate data flow names, so neither does the spec.

**ForceNew**: changing this destroys and recreates the data flow.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.flowlet

`bool`

Set true to create this data flow as a FLOWLET -- a reusable
snippet other data flows embed via their source/sink/
transformation `flowlet` references -- instead of a runnable
mapping data flow. Flowlets may omit sources and sinks (the
embedding flow supplies them). The form is the object's identity:
changing it replaces the data flow.

### spec.script

`string`

The data flow script as one string -- the transformation logic
from the Studio's "Script" view. Provide this or `script_lines`
(or both; Azure stores whichever is sent).

### spec.scriptLines

`[]string`

The data flow script as individual lines -- an alternative to
`script` that diffs better under version control. Provide this or
`script` (or both).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.sources

`[]AzureDataFactoryDataFlowSource`

The named input endpoints the script reads from. REQUIRED (at
least one) for a mapping data flow; optional for a flowlet. Each
source's `name` must match a stream name the script references.

### spec.sources[].name

`string` · required

The source's stream name -- what the script references.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.sources[].description

`string`

A human-readable description of this source.

### spec.sources[].dataset

`AzureDataFactoryDataFlowDatasetReference`

The dataset this source reads, by name.

### spec.sources[].dataset.name

`string | valueFrom` · required

The dataset's name inside the factory. A literal string or a
reference to an AzureDataFactoryDataset's name output.

- references: AzureDataFactoryDataset (`status.outputs.dataset_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryDataset, name: <that resource's name>, fieldPath: status.outputs.dataset_name}} -- a bare string does not parse

### spec.sources[].dataset.parameters

`map<string, string>`

Values for the dataset's parameters, if it declares any.

### spec.sources[].flowlet

`AzureDataFactoryDataFlowFlowletReference`

A flowlet embedded at this source, by name.

### spec.sources[].flowlet.name

`string | valueFrom` · required

The flowlet's name inside the factory -- defaults to referencing
an AzureDataFactoryDataFlow's data_flow_name output (a data flow
created with `flowlet: true`).

- references: AzureDataFactoryDataFlow (`status.outputs.data_flow_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryDataFlow, name: <that resource's name>, fieldPath: status.outputs.data_flow_name}} -- a bare string does not parse

### spec.sources[].flowlet.parameters

`map<string, string>`

Values for the flowlet's parameters, if it declares any.

### spec.sources[].flowlet.datasetParameters

`string`

Parameter values passed through to the flowlet's datasets, as the
raw expression string Azure expects.

### spec.sources[].linkedService

`AzureDataFactoryDataFlowLinkedServiceReference`

The linked service this source reads through, by name (for
dataset-less, inline sources).

### spec.sources[].linkedService.name

`string | valueFrom` · required

The linked service's name inside the factory. A literal string or
a reference to an AzureDataFactoryLinkedService's name output.

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.sources[].linkedService.parameters

`map<string, string>`

Values for the linked service's parameters, if it declares any.

### spec.sources[].schemaLinkedService

`AzureDataFactoryDataFlowLinkedServiceReference`

The linked service the source's schema is read from, by name
(schema drift scenarios where data and schema live apart).

### spec.sources[].schemaLinkedService.name

`string | valueFrom` · required

The linked service's name inside the factory. A literal string or
a reference to an AzureDataFactoryLinkedService's name output.

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.sources[].schemaLinkedService.parameters

`map<string, string>`

Values for the linked service's parameters, if it declares any.

### spec.sinks

`[]AzureDataFactoryDataFlowSink`

The named output endpoints the script writes to. REQUIRED (at
least one) for a mapping data flow; optional for a flowlet. Each
sink's `name` must match a stream name the script references.

### spec.sinks[].name

`string` · required

The sink's stream name -- what the script references.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.sinks[].description

`string`

A human-readable description of this sink.

### spec.sinks[].dataset

`AzureDataFactoryDataFlowDatasetReference`

The dataset this sink writes, by name.

### spec.sinks[].dataset.name

`string | valueFrom` · required

The dataset's name inside the factory. A literal string or a
reference to an AzureDataFactoryDataset's name output.

- references: AzureDataFactoryDataset (`status.outputs.dataset_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryDataset, name: <that resource's name>, fieldPath: status.outputs.dataset_name}} -- a bare string does not parse

### spec.sinks[].dataset.parameters

`map<string, string>`

Values for the dataset's parameters, if it declares any.

### spec.sinks[].flowlet

`AzureDataFactoryDataFlowFlowletReference`

A flowlet embedded at this sink, by name.

### spec.sinks[].flowlet.name

`string | valueFrom` · required

The flowlet's name inside the factory -- defaults to referencing
an AzureDataFactoryDataFlow's data_flow_name output (a data flow
created with `flowlet: true`).

- references: AzureDataFactoryDataFlow (`status.outputs.data_flow_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryDataFlow, name: <that resource's name>, fieldPath: status.outputs.data_flow_name}} -- a bare string does not parse

### spec.sinks[].flowlet.parameters

`map<string, string>`

Values for the flowlet's parameters, if it declares any.

### spec.sinks[].flowlet.datasetParameters

`string`

Parameter values passed through to the flowlet's datasets, as the
raw expression string Azure expects.

### spec.sinks[].linkedService

`AzureDataFactoryDataFlowLinkedServiceReference`

The linked service this sink writes through, by name (for
dataset-less, inline sinks).

### spec.sinks[].linkedService.name

`string | valueFrom` · required

The linked service's name inside the factory. A literal string or
a reference to an AzureDataFactoryLinkedService's name output.

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.sinks[].linkedService.parameters

`map<string, string>`

Values for the linked service's parameters, if it declares any.

### spec.sinks[].schemaLinkedService

`AzureDataFactoryDataFlowLinkedServiceReference`

The linked service the sink's schema is read from, by name
(schema drift scenarios where data and schema live apart).

### spec.sinks[].schemaLinkedService.name

`string | valueFrom` · required

The linked service's name inside the factory. A literal string or
a reference to an AzureDataFactoryLinkedService's name output.

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.sinks[].schemaLinkedService.parameters

`map<string, string>`

Values for the linked service's parameters, if it declares any.

### spec.sinks[].rejectedLinkedService

`AzureDataFactoryDataFlowLinkedServiceReference`

The linked service rejected rows are written through, by name.
SINKS ONLY: Azure's data flow model carries rejected-data routing
on sinks alone (the provider silently drops it on sources, so the
spec does not offer it there).

### spec.sinks[].rejectedLinkedService.name

`string | valueFrom` · required

The linked service's name inside the factory. A literal string or
a reference to an AzureDataFactoryLinkedService's name output.

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.sinks[].rejectedLinkedService.parameters

`map<string, string>`

Values for the linked service's parameters, if it declares any.

### spec.transformations

`[]AzureDataFactoryDataFlowTransformation`

Named intermediate transformations the script references (in
addition to what the script itself defines inline).

### spec.transformations[].name

`string` · required

The transformation's stream name -- what the script references.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.transformations[].description

`string`

A human-readable description of this transformation.

### spec.transformations[].dataset

`AzureDataFactoryDataFlowDatasetReference`

The dataset this transformation uses, by name.

### spec.transformations[].dataset.name

`string | valueFrom` · required

The dataset's name inside the factory. A literal string or a
reference to an AzureDataFactoryDataset's name output.

- references: AzureDataFactoryDataset (`status.outputs.dataset_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryDataset, name: <that resource's name>, fieldPath: status.outputs.dataset_name}} -- a bare string does not parse

### spec.transformations[].dataset.parameters

`map<string, string>`

Values for the dataset's parameters, if it declares any.

### spec.transformations[].flowlet

`AzureDataFactoryDataFlowFlowletReference`

A flowlet embedded at this transformation, by name.

### spec.transformations[].flowlet.name

`string | valueFrom` · required

The flowlet's name inside the factory -- defaults to referencing
an AzureDataFactoryDataFlow's data_flow_name output (a data flow
created with `flowlet: true`).

- references: AzureDataFactoryDataFlow (`status.outputs.data_flow_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryDataFlow, name: <that resource's name>, fieldPath: status.outputs.data_flow_name}} -- a bare string does not parse

### spec.transformations[].flowlet.parameters

`map<string, string>`

Values for the flowlet's parameters, if it declares any.

### spec.transformations[].flowlet.datasetParameters

`string`

Parameter values passed through to the flowlet's datasets, as the
raw expression string Azure expects.

### spec.transformations[].linkedService

`AzureDataFactoryDataFlowLinkedServiceReference`

The linked service this transformation uses, by name.

### spec.transformations[].linkedService.name

`string | valueFrom` · required

The linked service's name inside the factory. A literal string or
a reference to an AzureDataFactoryLinkedService's name output.

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.transformations[].linkedService.parameters

`map<string, string>`

Values for the linked service's parameters, if it declares any.

### spec.description

`string`

A human-readable description of what the data flow does.

### spec.annotations

`[]string`

Free-form annotation strings stored on the data flow.

### spec.folder

`string`

The folder the data flow appears under in the Data Factory
Studio, e.g. "transformations/daily". Omit for the root.

## Validation Rules

- `azure_data_factory_data_flow_script_required`: Provide the data flow script -- script, script_lines, or both
- `azure_data_factory_data_flow_mapping_source_sink`: A mapping data flow requires at least one source and one sink -- only flowlets (flowlet: true) may omit them

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDataFactoryDataFlow, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.data_flow_id` | `string` | The data flow's Azure Resource Manager ID ({factory_id}/dataflows/{name}) -- the same ID shape for both the mapping and flowlet forms. |
| `status.outputs.data_flow_name` | `string` | The data flow's name -- what other data flows' flowlet references resolve against. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.dataFactoryId` | AzureDataFactory | `status.outputs.data_factory_id` |
| `spec.sources[].dataset.name` | AzureDataFactoryDataset | `status.outputs.dataset_name` |
| `spec.sources[].flowlet.name` | AzureDataFactoryDataFlow | `status.outputs.data_flow_name` |
| `spec.sources[].linkedService.name` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.sources[].schemaLinkedService.name` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.sinks[].dataset.name` | AzureDataFactoryDataset | `status.outputs.dataset_name` |
| `spec.sinks[].flowlet.name` | AzureDataFactoryDataFlow | `status.outputs.data_flow_name` |
| `spec.sinks[].linkedService.name` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.sinks[].schemaLinkedService.name` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.sinks[].rejectedLinkedService.name` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.transformations[].dataset.name` | AzureDataFactoryDataset | `status.outputs.dataset_name` |
| `spec.transformations[].flowlet.name` | AzureDataFactoryDataFlow | `status.outputs.data_flow_name` |
| `spec.transformations[].linkedService.name` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureDataFactoryDataFlow | `spec.sources[].flowlet.name` | `status.outputs.data_flow_name` |
| AzureDataFactoryDataFlow | `spec.sinks[].flowlet.name` | `status.outputs.data_flow_name` |
| AzureDataFactoryDataFlow | `spec.transformations[].flowlet.name` | `status.outputs.data_flow_name` |

## See Also

- [Overview](../README.md)
