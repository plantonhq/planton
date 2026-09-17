# DigitalOcean Kafka Schema

Registers a schema subject (Avro, JSON Schema, or Protobuf) in a DigitalOcean managed Kafka cluster's schema registry, so producers and consumers agree on message structure. Every field is create-only: the provider has no update path, so any change to the definition destroys the subject and re-registers it, dropping all previously registered versions. Avro and JSON Schema definitions are rendered into the registry's canonical form before sending, so key order and whitespace in the manifest never count as a change. The owning cluster is wired by reference or supplied as a literal UUID and must run a General Purpose Kafka plan -- the registry exists only there.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Schema Registry Subject** -- the named subject on the referenced cluster's registry, carrying your schema definition

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Kafka Database Cluster** -- a DigitalOceanDatabaseCluster running the `kafka` engine on a General Purpose (dedicated-CPU: `gd-*`, `c2-*`, `m3-*`) plan. The schema registry exists only on those plans: a Basic-plan Kafka cluster answers every registry call `412 schema registry is disabled for this cluster` and cannot turn it on (`422 schema registry not supported for current plan`).

### DigitalOcean Account

- Nothing beyond the cluster: schema subjects are free API objects on it.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Kafka Schema**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Avro Event Schema** preset in the [Presets](#presets) tab to register a topic's value schema under the `<topic>-value` naming convention.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseKafkaSchema
metadata:
  name: orders-value-schema
  org: acme-corp
  env: prod
spec:
  cluster:
    value: "2f6a8f0e-3b1c-4c8e-9f2d-7a5b4c3d2e1f"
  subjectName: orders-value
  schemaType: avro
  schema: '{"type":"record","name":"Order","namespace":"com.acme.orders","fields":[{"name":"id","type":"string"},{"name":"amountCents","type":"long"}]}'
```

```shell
planton apply -f kafka-schema.yaml
```

This registers the `orders-value` subject on the referenced Kafka cluster's registry with a two-field Avro record as its founding schema. A Stack Job tracks the provisioning in real time.

### InfraChart

When the Kafka cluster deploys in the same InfraPipeline, wire the subject to it with ValueFromRef instead of a literal UUID:

```yaml
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: events-kafka
      fieldPath: status.outputs.cluster_id
```

The InfraPipeline resolves the dependency graph, deploys the cluster first, then registers the subject in its registry.

## Key Configuration

These are the most important decisions when configuring a Kafka schema subject. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Every change is a replacement** -- The provider has no update path: changing `schema`, `schemaType`, `subjectName`, or `cluster` destroys the subject and re-registers the new document as version 1, dropping every previously registered version. Consumers that pin older schema versions lose them the moment the replacement lands.

**Avro and JSON Schema are canonicalized; protobuf is not** -- The registry stores JSON schemas with object keys sorted and no whitespace, and the provider compares that stored text with yours verbatim. Both provisioners render Avro and JSON Schema definitions into the same canonical form before sending, so you may write the JSON in any key order and with any whitespace and a re-apply never proposes a change; only a real change (a field, a type, a default, a doc string) is a change. Protobuf text is sent as written, but the registry reformats it (a blank line after the `syntax` line, measured 2026-09-17) and the provider then sees a difference on every refreshed Terraform plan -- manage protobuf subjects with Pulumi, or author the text exactly as the registry renders it. See the GUIDE.

**Founding schema, not evolution channel** -- If your consumers rely on registry-mediated compatibility across versions, do not evolve schemas through this resource. Evolve them through your producers' registry client (which appends versions) and use this resource only to declare the founding schema of a subject.

**Subject naming** -- `subjectName` is the subject's API identity, unique within the cluster's registry. Follow the registry's `<topic>-value` (and `<topic>-key`) convention so serializer libraries resolve the schema automatically from the topic name.

**Schema language** -- `schemaType` accepts exactly `avro`, `json`, or `protobuf`, lowercase and case-sensitive. Changing it later replaces the subject.

**Compatibility level is not here** -- The registry's subject compatibility level (BACKWARD, FULL, and so on) has no surface in this resource; it stays whatever the registry defaults to or whatever was set out-of-band.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanDatabaseCluster** | `cluster` | `status.outputs.cluster_id` |

The referenced cluster must run the `kafka` engine; a literal cluster UUID is accepted in place of the reference.

### What This Component Provides

After provisioning, `status.outputs` carries only the subject's identity pair -- `cluster_id` and `subject_name` -- both echoes of resolved inputs. The registry's internal numeric schema id is discarded by the provider and deliberately not exported. Producers and consumers fetch the schema by subject name through the cluster's registry endpoint, authenticating with the cluster's connection outputs and user credentials -- there is no output here for downstream Cloud Resources to wire.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Avro value schema for an event topic** -- an Avro record registered under `<topic>-value` so producer and consumer serializers resolve it automatically. The default shape for event pipelines. Start from the **Avro Event Schema** preset.

**Strict JSON contract** -- a JSON Schema closed to unknown properties with enumerated fields, turning a loose JSON pipeline into a validated contract without moving to Avro. Start from the **JSON Contract Schema** preset.

## Works With

- [**DigitalOcean Database Cluster**](/cloud-catalog/digital-ocean-database-cluster) -- the Kafka-engine cluster whose registry holds the subject
- [**DigitalOcean Kafka Topic**](/cloud-catalog/digital-ocean-database-kafka-topic) -- the topic whose messages the subject describes, paired through the `<topic>-value` naming convention
- [**DigitalOcean Database User**](/cloud-catalog/digital-ocean-database-user) -- credentials producers and consumers use to reach the cluster and its registry
