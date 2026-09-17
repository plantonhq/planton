# DigitalOcean Kafka Topic

Creates a topic on a DigitalOcean managed Kafka cluster with the full per-topic configuration surface -- partitions, replication, cleanup policy, retention, and segment tuning. Partition count, replication factor, and every config leaf update in place, but Kafka only ever adds partitions -- lowering the count is rejected by the API. The owning cluster is wired by reference or supplied as a literal UUID.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Kafka Topic** -- the named topic on the referenced cluster, with your partition count, replication factor, and per-topic configuration; unset config leaves defer to the Kafka server default

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Kafka Database Cluster** -- a DigitalOceanDatabaseCluster running the `kafka` engine (DigitalOcean rejects topic calls on other engines).

### DigitalOcean Account

- Nothing beyond the cluster: topics are free API objects on it. Retained messages consume the cluster's paid storage -- retention settings are the cost lever.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Kafka Topic**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Event Stream Topic** preset in the [Presets](#presets) tab for a durable, time-retained event stream.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseKafkaTopic
metadata:
  name: orders-events
  org: acme-corp
  env: prod
spec:
  cluster:
    value: "2f6a8f0e-3b1c-4c8e-9f2d-7a5b4c3d2e1f"
  topicName: orders-events
  partitionCount: 6
  replicationFactor: 3
  config:
    cleanupPolicy: delete
    retentionMs: 604800000
    minInsyncReplicas: 2
```

```shell
planton apply -f kafka-topic.yaml
```

This creates an `orders-events` topic with six partitions, three replicas, and seven-day time-based retention on the referenced Kafka cluster. A Stack Job tracks the provisioning in real time.

### InfraChart

When the Kafka cluster deploys in the same InfraPipeline, wire the topic to it with ValueFromRef instead of a literal UUID:

```yaml
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: events-kafka
      fieldPath: status.outputs.cluster_id
```

The InfraPipeline resolves the dependency graph, deploys the cluster first, then creates the topic on it.

## Key Configuration

These are the most important decisions when configuring a Kafka topic. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Partitions are a one-way door** -- `partitionCount` accepts 3 to 2048; unset, DigitalOcean creates 3. Raising it applies in place, lowering it is an API error, and the only way down is destroying the topic and its messages. Adding partitions also reshuffles key-to-partition mapping for keyed producers, so consumers relying on key ordering see keys move. Size for the topic's mature throughput on day one. One more wrinkle: partition changes apply asynchronously and DigitalOcean never reports the live count back, so the manifest is the single source of truth -- out-of-band changes are invisible to drift detection.

**Renaming is deletion** -- `topicName` is the topic's API identity. Changing it destroys the topic -- messages, consumer offsets, everything -- and creates an empty one. Treat renames as migrations with a consumer cutover plan.

**Replication and durability travel together** -- `replicationFactor` takes at least 2 (DigitalOcean's default when unset); the ceiling is the cluster's node count, API-enforced. The floor for durable writes is `minInsyncReplicas: 2` with producers sending `acks=all` -- either alone protects nothing. Note `minInsyncReplicas` is the one config leaf the provider defaults locally (to 1) instead of reading the server value back, so leaving it unset always writes 1.

**Set `cleanupPolicy` explicitly whenever `config` is present** -- the provider seeds `cleanup_policy` to `compact_delete` the moment any config block exists, which is not Kafka's own `delete` default. If you add a config block just to tune retention, state the cleanup policy you actually want alongside it: `delete` discards old segments, `compact` keeps the latest value per key, `compact_delete` does both.

**Retention bounds the cluster's disk** -- `retentionMs` and `retentionBytes` accept -1 to mean no limit. On a compacted topic that is the normal shape (compaction bounds growth per key); on a `delete` topic it grows without bound against the cluster's paid storage. Set a real retention on every delete-policy topic.

**`messageFormatVersion` should never go backwards** -- consumers on older versions may not understand newer formats, and once raised the version should not be lowered. Leave it unset unless you are holding the format back for old consumers.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanDatabaseCluster** | `cluster` | `status.outputs.cluster_id` |

The referenced cluster must run the `kafka` engine; a literal cluster UUID is accepted in place of the reference.

### What This Component Provides

After provisioning, `status.outputs` carries the topic's identity pair -- `cluster_id` and `topic_name`, both echoes of resolved inputs. DigitalOcean mints no standalone topic id: the (cluster, name) pair is the identity. The topic's provisioning state is deliberately not an output: creation is asynchronous and a state captured at apply time goes stale, so anyone who needs it reads it live from the API (`GET /v2/databases/{cluster_id}/topics/{name}`), which is also how the E2E verifier asserts it. Producers and consumers connect through the cluster's connection outputs -- host, port, and credentials from its users -- and address the topic by name; there is no output here for downstream Cloud Resources to wire.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Durable event stream** -- seven-day time-based retention, three replicas with `minInsyncReplicas: 2` (pair with producers sending `acks=all`), and partition headroom for consumer parallelism. The default shape for event pipelines. Start from the **Event Stream Topic** preset.

**Compacted changelog** -- latest-value-per-key retention for state topics: pure compaction with an eager cleaner, a hard compaction deadline, and broker timestamps for consistent ordering. Start from the **Compacted Changelog Topic** preset.

## Works With

- [**DigitalOcean Database Cluster**](/cloud-catalog/digital-ocean-database-cluster) -- the Kafka-engine cluster the topic lives on, and the source of connection host, port, and credentials
- [**DigitalOcean Database User**](/cloud-catalog/digital-ocean-database-user) -- per-user Kafka ACLs pair this topic's name with produce/consume permissions
- [**DigitalOcean Kafka Schema**](/cloud-catalog/digital-ocean-database-kafka-schema) -- registers the message schema for the topic under the `<topic>-value` subject convention
