# GcpPubSubSchema

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpPubSubSchemaSpec defines a Pub/Sub schema — the message contract
publishers and subscribers agree on.

A schema is a first-class, shareable resource: one schema can validate
messages on many topics (each topic attaches it by reference via
schema_settings), so a platform team evolves the event contract in one
place. Attaching a schema to a topic makes Pub/Sub reject any published
message that does not conform — moving contract violations from the
consumer (where they surface as processing failures) to the publisher
(where the producing team can fix them).

Revision lifecycle (important):

  - The schema name is immutable; changing schema_name replaces the
    resource (and detaches nothing silently — topics reference schemas
    by name, so plan replacements deliberately).

  - Changing the definition does NOT replace the resource. It commits a
    new REVISION of the schema in place; existing revisions remain
    available and attached topics accept messages that conform to any
    available revision. A schema holds at most 20 revisions — beyond
    that, old revisions must be deleted manually before new commits
    succeed.

  - Deleting a schema while topics reference it leaves those topics
    validating against the sentinel "_deleted-schema_"; publishes then
    fail. Delete topic attachments first.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPubSubSchema
metadata:
  name: test-pubsub-schema
spec:
  projectId:
    value: my-gcp-project-123 # replace with your project ID
  schemaName: test-order-events
  # AVRO is the best default: human-readable, self-describing, and the
  # format Pub/Sub's BigQuery and Cloud Storage integrations understand.
  type: AVRO
  definition: |
    {
      "type": "record",
      "name": "OrderEvent",
      "fields": [
        {"name": "order_id", "type": "string"},
        {"name": "amount_cents", "type": "long"},
        {"name": "currency", "type": "string"}
      ]
    }
  # Explicit destroy behavior: DELETE removes the schema with the stack
  # (PREVENT would refuse; ABANDON would leave it serving unmanaged).
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.schemaName` | `string` | yes |  |  |
| `spec.type` | `string` | yes |  |  |
| `spec.definition` | `string` | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the schema will be created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.schemaName

`string` · required

Name of the Pub/Sub schema.
Must be 3-255 characters, start with a letter, and contain only letters,
numbers, hyphens, underscores, periods, tildes, plus signs, and percent
signs. Names beginning with "goog" are reserved by Google and rejected
at create time. Immutable after creation.

- rule: schema names beginning with 'goog' are reserved by Google — choose a different name
- rule: {"required":true,"string":{"minLen":"3","maxLen":"255","pattern":"^[a-zA-Z][a-zA-Z0-9\\-_\\.~+%]*$"}}

### spec.type

`string` · required

The schema definition language.
Valid values:
  - "AVRO": the definition is an Avro schema (JSON). Best default —
    human-readable, self-describing, and the format Pub/Sub's BigQuery
    and Cloud Storage subscription integrations understand natively.
  - "PROTOCOL_BUFFER": the definition is a protobuf message definition
    (a single message in proto2 or proto3 syntax). Choose when
    publishers already serialize protobuf and binary encoding matters.
Keep the type stable for the life of the schema: revisions must stay
compatible with the messages already validated against it, and topics
choose their encoding (JSON/BINARY) against this type.

- rule: type must be AVRO or PROTOCOL_BUFFER
- rule: {"required":true}

### spec.definition

`string` · required

The schema definition text.
For AVRO: a JSON Avro schema (e.g. {"type":"record","name":"Event",...}).
For PROTOCOL_BUFFER: a protobuf message definition (proto2/proto3 syntax).
Changing the definition commits a new schema REVISION in place (no
replacement); a schema holds at most 20 revisions, and revisions must
be backward-compatible with the encoding topics use, or publishers
pinned to older revisions will keep validating against those.

- rule: {"required":true}

### spec.deletionPolicy

`string`

Deletion policy for the schema — what happens when this resource is
destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the schema is deleted. Topics still referencing it fall
               back to the "_deleted-schema_" sentinel and their
               publishes start FAILING — detach topics first
  "PREVENT" -- destroy FAILS; protects a schema that many topics'
               validation depends on
  "ABANDON" -- the schema is removed from management but left serving
               in GCP (attached topics keep validating)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpPubSubSchema, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.schema_id` | `string` | Fully qualified schema resource path: projects/{project}/schemas/{name}. This is the exact string a topic's schema_settings.schema consumes — reference it from GcpPubSubTopic to attach message validation. |
| `status.outputs.schema_name` | `string` | The short name of the schema (same as the spec's schema_name input). Useful for display, logging, and human-readable references. |
| `status.outputs.revision_id` | `string` | The current revision ID of the schema (a new one is committed every time the definition changes). This is the exact value a topic's schema_settings.first_revision_id / last_revision_id consume — reference it from GcpPubSubTopic to pin validation to the revision this deploy produced. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpPubSubTopic | `spec.schemaSettings.schema` | `status.outputs.schema_id` |
| GcpPubSubTopic | `spec.schemaSettings.firstRevisionId` | `status.outputs.revision_id` |
| GcpPubSubTopic | `spec.schemaSettings.lastRevisionId` | `status.outputs.revision_id` |

## See Also

- [Overview](../README.md)
