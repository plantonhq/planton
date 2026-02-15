---
title: "Single Broker Kafka with UI"
description: "This preset deploys a single-broker Kafka cluster with a single ZooKeeper node, a sample topic, and the Kafka UI for visual monitoring. Ideal for development and testing."
type: "preset"
rank: "01"
presetSlug: "01-single-broker"
componentSlug: "kafka"
componentTitle: "Kafka"
provider: "kubernetes"
icon: "package"
order: 1
---

# Single Broker Kafka with UI

This preset deploys a single-broker Kafka cluster with a single ZooKeeper node, a sample topic, and the Kafka UI for visual monitoring. Ideal for development and testing.

## When to Use

- Development or testing of Kafka-based event-driven architectures
- Local or staging environments where a minimal Kafka setup is sufficient
- Prototyping with the Kafka UI for topic inspection and consumer group monitoring

## Key Configuration Choices

- **Single broker** with 10Gi disk -- sufficient for development; messages are stored on disk
- **Single ZooKeeper** -- minimal coordination; not HA
- **Kafka UI enabled** (`true`) -- web interface for topic browsing and consumer lag monitoring
- **Sample topic** (`events`) -- created with default partitions and replicas; add more topics as needed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **02-production-cluster** -- 3-broker, 3-ZooKeeper cluster for production
- **03-with-schema-registry** -- Adds Schema Registry for Avro/Protobuf schema management
