# DigitalOceanDatabaseKafkaSchema

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanDatabaseKafkaSchemaSpec models the full
digitalocean_database_kafka_schema_registry resource surface: one schema
subject registered in a DigitalOcean managed Kafka cluster's schema
registry.

EVERY field is create-only. There is no update path in the provider:
any change -- including evolving the schema definition -- destroys the
subject and re-registers it, which DROPS all previously registered
versions of the subject. Treat schema evolution as a deliberate
replacement, never a casual edit.

## Example

```yaml
# Reference manifests for DigitalOceanDatabaseKafkaSchema --
# protovalidate-valid, embedded as the reference page's Example block, and
# the documents the offline tofu plans render. Two documents: an Avro
# record subject and a JSON-schema subject.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseKafkaSchema
metadata:
  name: orders-value-schema
spec:
  # Literal cluster UUID; use valueFrom to reference a
  # DigitalOceanDatabaseCluster resource instead.
  cluster:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  subjectName: orders-value
  schemaType: avro
  schema: '{"type":"record","name":"Order","fields":[{"name":"id","type":"string"},{"name":"amountCents","type":"long"}]}'
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseKafkaSchema
metadata:
  name: shipment-events-schema
spec:
  cluster:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  subjectName: shipment-events-value
  schemaType: json
  schema: '{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","properties":{"shipmentId":{"type":"string"},"status":{"type":"string"}},"required":["shipmentId","status"]}'
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.cluster` | `string \| valueFrom` | yes |  | DigitalOceanDatabaseCluster (`status.outputs.cluster_id`) |
| `spec.subjectName` | `string` | yes |  |  |
| `spec.schemaType` | `string` | yes |  |  |
| `spec.schema` | `string` | yes |  |  |

## Field Details

### spec.cluster

`string | valueFrom` · required

The Kafka database cluster whose schema registry the subject is
registered in. Use a literal cluster UUID or a reference to a
DigitalOceanDatabaseCluster resource. The cluster must run the kafka
engine on a GENERAL PURPOSE (dedicated-CPU, `gd-*`/`c2-*`/`m3-*`) plan:
DigitalOcean's schema registry exists only there -- a Basic-plan Kafka
cluster answers every registry call "412 schema registry is disabled
for this cluster" and cannot enable it ("422 schema registry not
supported for current plan"), measured 2026-09-17. Changing the cluster
replaces the subject.

- references: DigitalOceanDatabaseCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDatabaseCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.subjectName

`string` · required

Name of the schema subject. Unique within the cluster's registry; the
subject name IS the API identity. Changing it replaces the subject.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.schemaType

`string` · required

Schema definition language: avro, json, or protobuf (case-sensitive).
Changing it replaces the subject.

- rule: {"required":true,"string":{"in":["avro","json","protobuf"]}}

### spec.schema

`string` · required

The schema definition itself, in the language schema_type names (e.g.
an Avro record document as JSON). Changing it replaces the subject and
drops all prior versions.

For avro and json, write the JSON in any key order and with any
whitespace: the registry stores JSON schemas in canonical form (object
keys sorted, no whitespace) and both provisioners render the definition
into that same form before sending, so what you write and what the
registry holds always agree and a re-apply never proposes a change. A
definition that is not valid JSON fails at plan time. For protobuf the
text is sent verbatim and the registry re-formats it on its side
(measured 2026-09-17: a blank line inserted after the `syntax` line),
so a protobuf subject re-plans a replacement on every refreshed
Terraform plan until the provider compares normalized text -- see the
GUIDE before managing protobuf subjects with Terraform.

- rule: {"required":true,"string":{"minLen":"1"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanDatabaseKafkaSchema, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_id` | `string` | UUID of the Kafka database cluster whose registry holds the subject. |
| `status.outputs.subject_name` | `string` | Name of the registered schema subject (its API identity within the cluster's registry). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cluster` | DigitalOceanDatabaseCluster | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
