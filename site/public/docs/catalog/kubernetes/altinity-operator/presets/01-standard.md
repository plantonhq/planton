---
title: "Standard Altinity ClickHouse Operator"
description: "This preset deploys the Altinity ClickHouse Operator with recommended default resources. The operator manages the lifecycle of ClickHouse clusters on Kubernetes, enabling declarative creation and..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "altinity-operator"
componentTitle: "Altinity Operator"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Altinity ClickHouse Operator

This preset deploys the Altinity ClickHouse Operator with recommended default resources. The operator manages the lifecycle of ClickHouse clusters on Kubernetes, enabling declarative creation and management of ClickHouse installations.

## When to Use

- You need to run ClickHouse databases on Kubernetes
- You want operator-managed ClickHouse cluster lifecycle (create, scale, backup, upgrade)
- Standard resource allocation is sufficient for the operator control plane

## Key Configuration Choices

- **Namespace** (`altinity-system`) -- dedicated namespace isolates the operator from workloads it manages
- **Create namespace** (`true`) -- namespace is created automatically if it does not exist
- **Resource requests** (`100m` CPU, `256Mi` memory) -- conservative baseline for the operator pod
- **Resource limits** (`1000m` CPU, `1Gi` memory) -- sufficient headroom for reconciliation spikes

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.
