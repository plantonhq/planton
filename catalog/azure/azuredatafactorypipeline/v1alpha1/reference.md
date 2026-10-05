# AzureDataFactoryPipeline

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureDataFactoryPipelineSpec** defines an Azure Data Factory
pipeline -- one unit of work inside a factory
(AzureDataFactory): an ordered set of activities (copy data, run a
data flow, call a stored procedure, wait, branch) that executes as
a whole when triggered.

The pipeline's activities travel as raw JSON (activities_json) --
the same "activities" array the Data Factory Studio's Code view
shows. Azure owns that schema (dozens of activity types, each with
its own shape); the catalog deliberately does not re-model it.

## Example

```yaml
# Deep-shape example for docs and offline validation: a pipeline with
# the full optional surface -- a Wait activity in the raw activities
# JSON, run parameters, variables, annotations, concurrency, a Studio
# folder, and the elapsed-time metric. References are literal values
# so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactoryPipeline
metadata:
  name: test-data-factory-pipeline
  id: test-data-factory-pipeline
  org: test-org
  env: test
spec:
  dataFactoryId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.DataFactory/factories/test-org-data-factory
  name: ingest-daily
  description: Copies one day's data into the lakehouse, parameterized by window start.
  parameters:
    windowStart: ""
  variables:
    cursor: ""
  activitiesJson: |
    [{"name": "placeholder-wait", "type": "Wait", "typeProperties": {"waitTimeInSeconds": 10}}]
  annotations:
    - team:data
    - tier:bronze
  concurrency: 1
  folder: ingest
  monitorMetricsAfterDuration: 0.00:30:00
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.dataFactoryId` | `string \| valueFrom` | yes |  | AzureDataFactory (`status.outputs.data_factory_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.parameters` | `map<string, string>` |  |  |  |
| `spec.variables` | `map<string, string>` |  |  |  |
| `spec.activitiesJson` | `string` |  |  |  |
| `spec.annotations` | `[]string` |  |  |  |
| `spec.concurrency` | `int32` |  |  |  |
| `spec.folder` | `string` |  |  |  |
| `spec.monitorMetricsAfterDuration` | `string` |  |  |  |

## Field Details

### spec.dataFactoryId

`string | valueFrom` · required

The Data Factory the pipeline lives in, by ARM ID. Can be a
literal string or a reference to an AzureDataFactory output.

**ForceNew**: changing this destroys and recreates the pipeline.

- references: AzureDataFactory (`status.outputs.data_factory_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactory, name: <that resource's name>, fieldPath: status.outputs.data_factory_id}} -- a bare string does not parse

### spec.name

`string` · required

The pipeline's name -- unique within the factory. Must start with
a letter, number, or underscore, and must not contain any of
< > * # . % & : \ + ? / (Azure's pipeline naming rules).

**ForceNew**: changing this destroys and recreates the pipeline.

- rule: Pipeline names must start with a letter, number, or underscore and must not contain < > * # . % & : \ + ? /
- rule: {"required":true}

### spec.description

`string`

A human-readable description of what the pipeline does.

### spec.parameters

`map<string, string>`

The pipeline's run-time parameters and their DEFAULT values --
callers override them per run. Azure types every parameter
declared here as String (declare other-typed parameters inside
activities_json if you need them; the provider round-trips only
string-typed entries).

### spec.variables

`map<string, string>`

The pipeline's variables and their default values -- mutable
state activities read and set during a run. The same
String-typing note as parameters applies.

### spec.activitiesJson

`string`

The pipeline's activities as a raw JSON ARRAY -- exactly the
"activities" array from the Data Factory Studio's Code view, e.g.
[{"name": "wait", "type": "Wait", "typeProperties":
{"waitTimeInSeconds": 10}}]. Azure owns the activity schema;
invalid JSON or unknown activity shapes fail at deploy time.
Key ordering inside the JSON is not meaningful (both engines
normalize it when diffing).

### spec.annotations

`[]string`

Free-form annotation strings stored on the pipeline.

### spec.concurrency

`int32` · optional (explicit presence)

The maximum number of concurrent runs of this pipeline, 1-50.
Omit for Azure's unlimited default.

- rule: {"int32":{"lte":50,"gte":1}}

### spec.folder

`string`

The folder the pipeline appears under in the Data Factory
Studio, e.g. "ingest/daily". Omit for the root.

### spec.monitorMetricsAfterDuration

`string`

How long the pipeline may run before its elapsed-time metric
fires, as a Data Factory TimeSpan string (e.g. "0.00:30:00" for
30 minutes). Omit to skip the metric.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDataFactoryPipeline, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.pipeline_id` | `string` | The pipeline's Azure Resource Manager ID ({factory_id}/pipelines/{name}). |
| `status.outputs.pipeline_name` | `string` | The pipeline's name -- what triggers and pipeline-run API calls reference. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.dataFactoryId` | AzureDataFactory | `status.outputs.data_factory_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureDataFactoryTrigger | `spec.schedule.pipelines[].name` | `status.outputs.pipeline_name` |
| AzureDataFactoryTrigger | `spec.tumblingWindow.pipeline.name` | `status.outputs.pipeline_name` |
| AzureDataFactoryTrigger | `spec.blobEvent.pipelines[].name` | `status.outputs.pipeline_name` |
| AzureDataFactoryTrigger | `spec.customEvent.pipelines[].name` | `status.outputs.pipeline_name` |

## See Also

- [Overview](../README.md)
