# GCP Managed Kafka Connector

Declares one data pipeline between Kafka and another system -- Pub/Sub, BigQuery, Cloud Storage, or another Kafka cluster -- in the manifest of the team whose data it moves, running on a shared Connect cluster with automatic task restarts.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Connector** -- a `managed_kafka_connector` on the referenced Connect cluster

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Managed Service for Apache Kafka admin permissions (`roles/managedkafka.admin`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpManagedKafkaConnectCluster`** -- the Connect cluster that runs the connector (`connectCluster`).

### Optional Dependencies

- **`GcpManagedKafkaTopic`** -- the topics a sink reads or a source writes, named in `configs`.
- **The destination or source** -- a Pub/Sub topic, BigQuery dataset, or bucket, with the Managed Kafka service agent granted access to it.

## Deploy

### Console

Open the deployment store, find **GCP Managed Kafka Connector**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Pub/Sub Sink** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaConnector
metadata:
  name: orders-to-pubsub
  org: acme-corp
  env: prod
spec:
  location: us-central1
  connectCluster:
    value: events-connect
  configs:
    connector.class: com.google.pubsub.kafka.sink.CloudPubSubSinkConnector
    tasks.max: "3"
    topics: orders
    cps.project: acme-prod
    cps.topic: orders-events
  taskRestartPolicy:
    minimumBackoff: 60s
    maximumBackoff: 1800s
```

```shell
planton apply -f managed-kafka-connector.yaml
```

This streams every message on the `orders` Kafka topic into the `orders-events` Pub/Sub topic. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference the `GcpManagedKafkaConnectCluster` from `connectCluster`; name the `GcpManagedKafkaTopic` it reads in `configs.topics`.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The plugin** -- `configs.connector.class` picks what the connector does; the rest of `configs` is that plugin's own configuration.

**Parallelism** -- `tasks.max` spreads the connector across Connect workers; a sink can use at most one task per partition.

**Recovery** -- `taskRestartPolicy` restarts failed tasks with exponential backoff between the two durations.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpManagedKafkaConnectCluster** | `connectCluster` | `status.outputs.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The connector's full resource name | Audit and console links |
| `connector_id` | The id | Operations |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Kafka to Pub/Sub** -- Google's Pub/Sub sink connector publishing a topic's messages. Start from the **Pub/Sub Sink** preset.

**Kafka to BigQuery** -- JSON messages streamed into BigQuery tables. Start from the **BigQuery Sink** preset.

**Cluster replication** -- Topics replicated from another Kafka cluster. Start from the **MirrorMaker 2 Source** preset.

## Works With

- [**GCP Managed Kafka Connect Cluster**](/cloud-catalog/gcp-managed-kafka-connect-cluster) -- the workers that run it
- [**GCP Managed Kafka Topic**](/cloud-catalog/gcp-managed-kafka-topic) -- the topics it reads or writes
- [**GCP Pub/Sub Topic**](/cloud-catalog/gcp-pub-sub-topic) -- a sink or source destination
