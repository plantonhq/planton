---
title: "Standard Strimzi Kafka Operator"
description: "This preset deploys the Strimzi Kafka Operator with recommended default resources. Strimzi provides a way to run Apache Kafka on Kubernetes using custom resources for Kafka clusters, topics, users,..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "strimzi-kafka-operator"
componentTitle: "Strimzi Kafka Operator"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Strimzi Kafka Operator

This preset deploys the Strimzi Kafka Operator with recommended default resources. Strimzi provides a way to run Apache Kafka on Kubernetes using custom resources for Kafka clusters, topics, users, and connectors.

## When to Use

- You need to run Apache Kafka on Kubernetes
- You want operator-managed Kafka cluster lifecycle (provisioning, scaling, rolling upgrades, topic management)
- Standard resource allocation is sufficient for the operator control plane

## Key Configuration Choices

- **Namespace** (`strimzi-system`) -- dedicated namespace keeps the operator separate from Kafka workloads
- **Create namespace** (`true`) -- namespace is created automatically if it does not exist
- **Resource requests** (`50m` CPU, `100Mi` memory) -- lightweight baseline for the operator pod
- **Resource limits** (`1000m` CPU, `1Gi` memory) -- sufficient headroom for managing multiple Kafka clusters

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.
