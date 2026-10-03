# GCP Pub/Sub Schema

Deploys a Pub/Sub schema (`google_pubsub_schema`) — the message contract publishers and subscribers agree on. One schema can validate messages on many topics: each topic attaches it by reference, so a platform team evolves the event contract in one place.

## Overview

A schema is a first-class project-level resource holding a typed message definition — an Avro record (JSON) or a protobuf message. A `GcpPubSubTopic` attaches it through `schemaSettings.schema` together with an encoding (`JSON` or `BINARY`); from that point Pub/Sub rejects any published message that does not conform. That moves contract violations from consumers (where they surface as processing failures deep in a pipeline) to publishers (where the producing team can fix them immediately).

Three properties define the operational model:

- **Shared** — many topics, one schema; the contract is evolved in one place.
- **Revisioned, not replaced** — changing the `definition` commits a new schema revision in place (a schema retains up to 20). Only renaming the schema replaces the resource.
- **Delete-ordered** — deleting a schema while topics reference it leaves those topics validating against the `_deleted-schema_` sentinel, and publishes fail. Detach or recreate topics first.

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPubSubSchema
metadata:
  name: order-events-schema
spec:
  schemaName: order-events
  type: AVRO
  definition: |
    {
      "type": "record",
      "name": "OrderEvent",
      "fields": [
        {"name": "order_id", "type": "string"},
        {"name": "amount_cents", "type": "long"}
      ]
    }
```

```shell
planton apply -f schema.yaml
```

Then attach it from a topic:

```yaml
# In a GcpPubSubTopic spec:
schemaSettings:
  schema:
    valueFrom:
      kind: GcpPubSubSchema
      name: order-events-schema
      fieldPath: status.outputs.schema_id
  encoding: JSON
```

## Configuration Options

| Category | Options |
|----------|---------|
| **Identity** | `schemaName` — 3-255 chars, starts with a letter; names beginning with `goog` are reserved by Google; immutable. `projectId` — optional; omitted rides the provider default; reference `GcpProject` |
| **Contract language** | `type` — `AVRO` (JSON Avro schema; best default) or `PROTOCOL_BUFFER` (a single proto2/proto3 message) |
| **Definition** | `definition` — the schema text; changing it commits a new revision in place |
| **Lifecycle** | `deletionPolicy` — `DELETE` (default), `PREVENT` (destroy fails; protects a schema many topics validate against), or `ABANDON` (remove from management, leave serving) |

### Choosing a type

`AVRO` reviews well in pull requests, carries logical types (timestamps, decimals), and is the format Pub/Sub's BigQuery delivery (`useTopicSchema`) and Cloud Storage Avro export understand natively. `PROTOCOL_BUFFER` suits publishers that already serialize protobuf and want compact binary encoding (pair with a topic encoding of `BINARY`).

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `schema_id` | string | Fully qualified schema path (`projects/{p}/schemas/{name}`) — the exact string a topic's `schemaSettings.schema` reference consumes |
| `schema_name` | string | The short name of the schema |
| `revision_id` | string | The current revision ID (a new one is committed on every definition change) — the exact value a topic's `schemaSettings.firstRevisionId`/`lastRevisionId` consume for revision pinning |

## Important Notes

- **Revision limit**: a schema retains at most 20 revisions. Beyond that, commits fail until old revisions are deleted manually (`gcloud pubsub schemas delete-revision`).
- **Keep revisions compatible**: attached topics accept messages conforming to any available revision; an incompatible new revision silently widens (or breaks) what publishers can send. Evolve additively.
- **Keep the type stable**: revisions should keep the declared type; a type flip mid-life invalidates existing publishers.

### Provider parity

Every configurable argument of `google_pubsub_schema` at the pinned
provider version is representable through this spec; the recorded
judgment lives in `iac/provider-parity.yaml`, checked by
`planton provider-parity --check`.

## Related Kinds

- **GcpPubSubTopic** — attaches this schema via `schemaSettings.schema`
- **GcpPubSubSubscription** — consumes schema-validated messages (BigQuery delivery's `useTopicSchema` and Cloud Storage `avroConfig.useTopicSchema` derive layout from the topic's schema)
- **GcpProject** — provides the GCP project ID

## Additional Resources

- [Pub/Sub schemas](https://cloud.google.com/pubsub/docs/schemas)
- [Schemas REST API](https://cloud.google.com/pubsub/docs/reference/rest/v1/projects.schemas)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
