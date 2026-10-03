# Kafka Topic

Declares ONE Kafka topic on a Strimzi-managed cluster. The declaration renders a `KafkaTopic` custom resource, and the target cluster's own TOPIC OPERATOR (enabled by default on Apache Kafka) reconciles it into a real topic — creating it, growing partitions, and applying configuration changes declaratively. The placement contract is strict and worth internalizing up front: the KafkaTopic must live in the SAME NAMESPACE as its Kafka cluster and name that cluster through the `strimzi.io/cluster` label (rendered from `kafkaCluster`) — a topic in another namespace, or naming a cluster that does not exist there, is accepted by the API server and then silently never reconciled. Deleting this resource deletes the TOPIC AND ITS DATA (the topic operator propagates deletion to Kafka).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **KafkaTopic** -- the Strimzi custom resource declaring the topic (partitions, replication factor, topic-level configuration), labeled `strimzi.io/cluster: <kafka_cluster>`
- **Kafka topic** (created by the cluster's topic operator, not the module) -- the real topic producers and consumers use, named `topic_name` when set, otherwise this resource's name

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **A Kafka cluster** -- an Apache Kafka resource (or an existing Strimzi `Kafka`) running on the cluster, with its entity operator's topic operator enabled (the default).
- The namespace you declare must be the Kafka cluster's OWN namespace — its topic operator watches only there.
- The replication factor you declare (if any) must not exceed the cluster's broker count; the topic operator rejects a higher value at reconcile time (the resource reports NotReady, nothing is created).

## Deploy

### Console

Open the deployment store, find **Kafka Topic**, and click **Deploy**. The creation wizard walks you through the cluster + namespace placement pair, the optional Kafka-side name override, partitions and replication factor, and the topic configuration map. Start from the **Simple Event Stream** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaTopic
metadata:
  name: order-events
  org: acme-corp
  env: prod
spec:
  namespace:
    value: kafka
  kafkaCluster:
    value: prod-kafka
  partitions: 6
  replicas: 3
  config:
    retention.ms: "604800000"
    cleanup.policy: delete
    min.insync.replicas: "2"
```

```shell
planton apply -f kafka-topic.yaml
```

This declares a 6-partition, replication-factor-3 event topic retained for 7 days — durable through one broker loss for producers using acks=all. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire both placement facts from the Kafka cluster itself so the pair can never drift apart:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesKafka
      name: prod-kafka
      fieldPath: status.outputs.namespace
  kafkaCluster:
    valueFrom:
      kind: KubernetesKafka
      name: prod-kafka
      fieldPath: status.outputs.cluster_name
  partitions: 6
  replicas: 3
```

The InfraPipeline deploys the Kafka cluster first, then declares the topic against it.

## Key Configuration

These are the most important decisions when configuring the topic. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Same-namespace placement is a contract, not a preference** -- `namespace` must be the Kafka cluster's own namespace. The cluster's topic operator watches only there; a topic anywhere else is accepted and silently never reconciled. Reference the Apache Kafka resource in `kafkaCluster` to inherit its `cluster_name` output and draw the dependency edge.

**Plan partitions up front** -- empty inherits the cluster's `num.partitions` default. Partitions can be INCREASED later but never decreased (Kafka has no partition shrink), and increasing them changes key-to-partition mapping for keyed topics — semantic partitioning deserves a deliberate count from day one.

**Replication pairs with min.insync.replicas** -- empty inherits the cluster's `default.replication.factor`. The production norm is `replicas: 3` with `min.insync.replicas: "2"` in `config`: one broker can be lost without losing acknowledged writes and without halting producers. The factor cannot exceed the cluster's broker count.

**Topic name override only when Kubernetes cannot carry the name** -- empty means the resource's name IS the topic name. Set `topic_name` when the Kafka name needs `.`, `_`, or uppercase (e.g. `orders.v1_DLQ`); names are limited to 249 characters of alphanumerics, `.`, `_` and `-`. Names differing only by `.` vs `_` collide in Kafka's internal metrics — avoid coexisting pairs.

**Config values are Kafka configuration strings** -- write numbers and booleans as strings (`"604800000"`, `"false"`). Entries not declared inherit the cluster's broker-level defaults; changing an entry later reconfigures the LIVE topic declaratively.

**Deletion destroys data** -- deleting this resource deletes the topic and everything in it; the topic operator propagates the deletion to Kafka.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesKafka** | `kafkaCluster` | `status.outputs.cluster_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the KafkaTopic resource lives in (the Kafka cluster's namespace) | Placing related resources beside the cluster |
| `topic_name` | The actual Kafka topic name (`spec.topic_name` when set, otherwise `metadata.name`) | What producers and consumers subscribe to — wire workload env/config to THIS. The bootstrap endpoint comes from the Apache Kafka resource's own outputs |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Application event stream** -- 6 partitions, replication 3, 7-day delete retention. Start from the **Simple Event Stream** preset.

**Changelog / snapshot table** -- `cleanup.policy: compact` keeps the latest value per key; records must be keyed. Start from the **Compacted Changelog** preset.

**Telemetry firehose** -- 24 partitions, 2-day retention, 1 GiB segments for high-volume ingest. Start from the **High Throughput** preset.

## Works With

- [**Apache Kafka**](/infra-catalog/kubernetes-kafka) -- the cluster this topic belongs to; its topic operator does the reconciling, and its outputs carry the bootstrap endpoint clients connect to
- [**Kafka User**](/infra-catalog/kubernetes-kafka-user) -- authenticated principals with ACLs on this topic; declared the same way and reconciled by the same cluster's user operator
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- producers and consumers; reference this resource's `topic_name` output instead of hardcoding the name (the same wiring applies to StatefulSets and CronJobs)
