---
title: "Kafka Cluster with Schema Registry"
description: "This preset deploys a 3-broker Kafka cluster with the Confluent Schema Registry enabled. The Schema Registry provides centralized schema management for Avro, Protobuf, and JSON Schema serialization..."
type: "preset"
rank: "03"
presetSlug: "03-with-schema-registry"
componentSlug: "kafka"
componentTitle: "Kafka"
provider: "kubernetes"
icon: "package"
order: 3
---

# Kafka Cluster with Schema Registry

This preset deploys a 3-broker Kafka cluster with the Confluent Schema Registry enabled. The Schema Registry provides centralized schema management for Avro, Protobuf, and JSON Schema serialization formats.

## When to Use

- Event-driven systems using Avro, Protobuf, or JSON Schema for message serialization
- Teams that need schema evolution and compatibility enforcement across producers and consumers
- Production Kafka clusters where schema governance is required

## Key Configuration Choices

- **Schema Registry enabled** with 1 replica -- provides an HTTP API for schema registration and retrieval
- **3 brokers, 3 ZooKeeper** -- production-grade Kafka cluster (same as 02-production-cluster)
- **Kafka UI enabled** -- includes Schema Registry integration for visual schema management

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **01-single-broker** -- Minimal Kafka for development
- **02-production-cluster** -- Production cluster without Schema Registry
