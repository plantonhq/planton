# GCP Managed Kafka Topic

One topic on a Managed Service for Apache Kafka cluster -- its partitions, replication, and topic-level configuration. Topics are their own block so the team that produces to a topic owns it in its own manifest, next to the service that writes to it, without editing the cluster.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kafka topic** -- a `managed_kafka_topic` on the referenced cluster

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Managed Service for Apache Kafka admin permissions (`roles/managedkafka.admin`) on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpManagedKafkaCluster`** -- the cluster the topic lives on (`cluster`), by reference or by name.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaTopic
metadata:
  name: orders
spec:
  location: us-central1
  cluster:
    valueFrom:
      kind: GcpManagedKafkaCluster
      name: events
      fieldPath: status.outputs.name
  partitionCount: 12
  replicationFactor: 3
```

```shell
planton apply -f managed-kafka-topic.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The cluster's region. Immutable. |
| `cluster` | `StringValueOrRef` | The cluster -- a `GcpManagedKafkaCluster` reference or its full path or bare id. Immutable. |
| `partitionCount` | `int32` | At least 1. Grows in place, never shrinks. |
| `replicationFactor` | `int32` | Copies of each partition; Google recommends 3. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The cluster's project. |
| `topicId` | `string` | `metadata.name` | The topic name clients use (Kafka's rule: letters, digits, `.`, `_`, `-`, at most 249). Immutable. |
| `configs` | `map<string,string>` | cluster defaults | Topic-level overrides (`cleanup.policy`, `retention.ms`, ...). Replaced whole: a removed key returns to the cluster default. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `replicationFactor` and `partitionCount` are each at least 1.
- Every `configs` value is non-empty.
- A literal full cluster path must name this topic's `location` (and its `projectId`, when set).
- `topicId` follows Kafka's topic-name rule.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/clusters/{cluster}/topics/{topic_id}` |
| `topic_id` | `string` | The topic name Kafka clients use |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Replacement deletes messages.** `projectId`, `location`, `topicId`, `replicationFactor`, and the cluster are immutable -- changing one replaces the topic and its data.
- **Partitions only grow.** Google refuses a lower `partitionCount` at apply. Raising `partitionCount` moves keys to different partitions for new messages, so per-key ordering holds only after the change.
- **Access is a separate block.** Grant producers and consumers with `GcpManagedKafkaAcl` on `topic/{name}` (and the consumer group).

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpManagedKafkaCluster** -- the cluster the topic lives on
- **GcpManagedKafkaAcl** -- who may read and write the topic
- **GcpManagedKafkaConnector** -- pipelines that read or write it

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
