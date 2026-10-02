# GCP Managed Kafka Topic

Declares one Kafka topic in the manifest of the team that owns it: its partitions, replication, and retention or compaction settings, on a shared Managed Kafka cluster.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Kafka topic** -- a `managed_kafka_topic` on the referenced cluster

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Managed Service for Apache Kafka admin permissions (`roles/managedkafka.admin`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpManagedKafkaCluster`** -- the cluster the topic lives on (`cluster`), by reference or by name.

## Deploy

### Console

Open the deployment store, find **GCP Managed Kafka Topic**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Event Stream** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaTopic
metadata:
  name: orders
  org: acme-corp
  env: prod
spec:
  location: us-central1
  cluster:
    value: events
  partitionCount: 12
  replicationFactor: 3
```

```shell
planton apply -f managed-kafka-topic.yaml
```

This creates the `orders` topic with 12 partitions, three replicas each, on the `events` cluster. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference the `GcpManagedKafkaCluster` from `cluster` so the topic deploys after the cluster; pair it with a `GcpManagedKafkaAcl` for the services that use it.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Partitions** -- `partitionCount` is the unit of consumer parallelism; it can grow later but never shrink.

**Replication** -- `replicationFactor: 3` keeps a copy in each of the cluster's three zones -- the high-availability choice. It cannot change after creation.

**Retention and compaction** -- `configs` overrides the cluster's defaults per topic: `retention.ms` for time-based retention, `cleanup.policy: compact` to keep the latest value per key.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpManagedKafkaCluster** | `cluster` | `status.outputs.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The topic's full resource name | Audit and console links |
| `topic_id` | The topic name | A connector's `topics` config, client configuration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Event stream** -- 12 partitions, three replicas, seven days of retention. Start from the **Event Stream** preset.

**Changelog** -- A log-compacted topic that keeps the latest value per key. Start from the **Compacted Changelog** preset.

## Works With

- [**GCP Managed Kafka Cluster**](/cloud-catalog/gcp-managed-kafka-cluster) -- the cluster the topic lives on
- [**GCP Managed Kafka ACL**](/cloud-catalog/gcp-managed-kafka-acl) -- access for producers and consumers
- [**GCP Managed Kafka Connector**](/cloud-catalog/gcp-managed-kafka-connector) -- pipelines in and out of the topic
