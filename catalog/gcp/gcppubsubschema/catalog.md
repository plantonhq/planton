# GCP Pub/Sub Schema

Deploys a Pub/Sub schema — the message contract publishers and subscribers agree on. A schema is a first-class, shareable resource: one schema can validate messages on many topics (each topic attaches it by reference), so a platform team evolves the event contract in one place. Attaching a schema to a topic makes Pub/Sub reject any published message that does not conform, moving contract violations from the consumer to the publisher. Definition edits commit in-place revisions rather than replacing the resource, which is how contracts evolve without a coordinated publisher deploy.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Pub/Sub API enablement** -- the module enables `pubsub.googleapis.com` in the target project before creating the schema (and never disables it on destroy, so tearing down one schema cannot break the project's other Pub/Sub resources)
- **Pub/Sub Schema** -- a named schema resource in the specified GCP project, holding the contract definition in Avro or Protocol Buffers form
- **Initial Revision** -- the first committed revision of the definition; later definition edits commit additional in-place revisions (up to 20 kept per schema)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the schema will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The module enables the Pub/Sub API itself — no manual API setup is needed.

## Deploy

### Console

Open the deployment store, find **GCP Pub/Sub Schema**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **AVRO Event Contract** preset in the [Presets](#presets) tab to pre-populate a minimal configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPubSubSchema
metadata:
  name: order-event-contract
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  schemaName: order-event-contract
  type: AVRO
  definition: |
    {
      "type": "record",
      "name": "OrderCreated",
      "fields": [
        { "name": "orderId", "type": "string" },
        { "name": "amountCents", "type": "long" }
      ]
    }
```

```shell
planton apply -f pubsub-schema.yaml
```

This creates the schema with its first revision committed. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the schema to a GCP project deployed in the same InfraPipeline — and wire topics to the schema:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph, deploys the project first, then provisions the schema — and any GcpPubSubTopic referencing this schema's `schema_id` output deploys after it.

## Key Configuration

These are the most important decisions when configuring a Pub/Sub schema. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Schema name** -- The `schemaName` field is immutable; changing it REPLACES the resource, and topics reference schemas by name, so plan replacements deliberately across every attached topic.

**Contract language** -- The `type` field (`AVRO` or `PROTOCOL_BUFFER`) is stable for the schema's lifetime. Avro is the common choice for event contracts and analytics pipelines; Protocol Buffers suits fleets already speaking protobuf.

**Definition and revisions** -- Changing `definition` does NOT replace the resource: it commits a new in-place REVISION. Attached topics accept messages conforming to any available revision, which is how contracts evolve without a coordinated publisher deploy. A schema holds at most 20 revisions — beyond that, old revisions must be deleted before new commits succeed.

**Deletion order** -- Deleting a schema while topics still reference it leaves those topics validating against the `_deleted-schema_` sentinel, and every publish fails. Detach or destroy topics first. For a schema many topics depend on, set `deletionPolicy: PREVENT` so destroy fails instead of breaking publishers; `ABANDON` removes the schema from management but leaves it serving in GCP.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `schema_id` | Fully qualified schema ID (`projects/{project}/schemas/{name}`) | GcpPubSubTopic `schemaSettings.schema` — attaches publish-time validation |
| `revision_id` | The revision committed by this deploy (a new one per definition change) | GcpPubSubTopic `schemaSettings.firstRevisionId` / `lastRevisionId` — pins topic validation to a known revision |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Avro event contract** -- A JSON-defined record schema validating a domain event stream. The default posture for event-driven architectures. Start from the **AVRO Event Contract** preset.

**Protobuf binary contract** -- A proto3 message definition for services already speaking protobuf end to end, paired with BINARY encoding on the attaching topic. Start from the **Protobuf Binary Contract** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the schema is created
- [**GCP Pub/Sub Topic**](/infra-catalog/gcp-pub-sub-topic) -- attaches this schema to enforce the contract at publish time
