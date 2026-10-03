# Kubernetes Kafka Topic

## When NOT to Use This

**If the topic is auto-created by applications, leave it out of the catalog — or better, declare it and turn auto-creation off.** Kafka brokers create topics automatically by default (`auto.create.topics.enable`), with default configuration, when a client produces to or consumes from a name that does not exist. The topic operator manages ONLY topics that have a KafkaTopic resource; topics created and managed directly in Kafka are left alone. But mixing the two models is hazardous: an auto-created topic that races a KubernetesKafkaTopic declaration lands with default configuration first, and the operator's follow-up reconfiguration can fail (e.g. the declared partition count is lower than the default). Declared topics plus `auto.create.topics.enable: "false"` on the cluster is the recommended posture.

## Overview

**KubernetesKafkaTopic** declares ONE Kafka topic on the Strimzi `KafkaTopic` custom resource. The target cluster's TOPIC OPERATOR (enabled by default on KubernetesKafka) reconciles it into a real topic — creating it, growing partitions, and applying configuration changes declaratively. Configuration flows one way: changes made to the topic outside the resource are reverted at the next reconciliation.

The contract worth internalizing before the first apply:

- **Placement** — the KafkaTopic must live in the SAME NAMESPACE as its Kafka cluster and binds to it through the `strimzi.io/cluster` label (rendered from `kafka_cluster`). The topic operator watches only that namespace; a topic anywhere else is accepted by the API server and then silently never reconciled
- **Partitions grow, never shrink** — Kafka has no partition shrink; a decrease is rejected at reconcile (`PartitionDecreaseException`, resource NotReady). Increasing partitions changes key-to-partition mapping for keyed topics — plan counts up front for topics with semantic partitioning
- **Replicas are capped by broker count** — a replication factor above the cluster's broker count is rejected at reconcile time (the resource reports NotReady, nothing is created). 3 is the production norm, paired with `min.insync.replicas: "2"`
- **Deletion deletes the topic AND ITS DATA** — the topic operator propagates deletion to Kafka
- **`topic_name` carries names Kubernetes metadata cannot** — Kafka allows `.`, `_` and uppercase (e.g. `orders.v1_DLQ`); Kubernetes resource names do not. Empty = the resource's `metadata.name`

## Essential Configuration Fields

### Required

- **`spec.namespace`**: MUST be the Kafka cluster's own namespace (the placement contract above)
- **`spec.kafka_cluster`**: the cluster this topic belongs to — a literal cluster name or a reference to a KubernetesKafka resource; rendered as the `strimzi.io/cluster` label

### Common

- **`spec.partitions`**: empty = the cluster's `num.partitions` default
- **`spec.replicas`**: empty = the cluster's `default.replication.factor`
- **`spec.config`**: Kafka topic-level entries (`retention.ms`, `cleanup.policy`, `max.message.bytes`, `min.insync.replicas`, ...) — values are Kafka configuration strings, so write numbers and booleans as strings

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Where the KafkaTopic resource lives (the Kafka cluster's namespace) |
| `topic_name` | The actual Kafka topic name (`spec.topic_name` when set, otherwise `metadata.name`) — what producers and consumers subscribe to |

## Composing in Infra Charts

`KubernetesKafka → KubernetesKafkaTopic → workload` deploys in one chart run: the topic references the cluster's `cluster_name` output, and workloads combine this resource's `topic_name` output with the cluster's `internal_bootstrap_endpoint` output to configure clients. The bootstrap endpoint deliberately does not live here — it belongs to the cluster.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
