---
title: "Presets"
description: "Ready-to-deploy configuration presets for Kafka"
type: "preset-list"
componentSlug: "kafka"
componentTitle: "Kafka"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-single-broker"
    rank: "01"
    title: "Single Broker Kafka with UI"
    excerpt: "This preset deploys a single-broker Kafka cluster with a single ZooKeeper node, a sample topic, and the Kafka UI for visual monitoring. Ideal for development and testing."
  - slug: "02-production-cluster"
    rank: "02"
    title: "Production Kafka Cluster"
    excerpt: "This preset deploys a 3-broker, 3-ZooKeeper Kafka cluster with production-grade resources, replication, and the Kafka UI. Provides fault tolerance and horizontal throughput scaling."
  - slug: "03-with-schema-registry"
    rank: "03"
    title: "Kafka Cluster with Schema Registry"
    excerpt: "This preset deploys a 3-broker Kafka cluster with the Confluent Schema Registry enabled. The Schema Registry provides centralized schema management for Avro, Protobuf, and JSON Schema serialization..."
---

# Kafka Presets

Ready-to-deploy configuration presets for Kafka. Each preset is a complete manifest you can copy, customize, and deploy.
