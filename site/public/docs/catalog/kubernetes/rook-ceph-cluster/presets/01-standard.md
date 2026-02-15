---
title: "Standard Rook Ceph Cluster"
description: "This preset creates a Ceph cluster managed by the Rook operator with the toolbox and dashboard enabled. Requires the `KubernetesRookCephOperator` to be deployed first in the same namespace."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "rook-ceph-cluster"
componentTitle: "Rook Ceph Cluster"
provider: "kubernetes"
icon: "package"
order: 1
---

# Standard Rook Ceph Cluster

This preset creates a Ceph cluster managed by the Rook operator with the toolbox and dashboard enabled. Requires the `KubernetesRookCephOperator` to be deployed first in the same namespace.

## When to Use

- You need distributed storage (block, file, or object) on Kubernetes
- The Rook Ceph Operator is already running in the `rook-ceph` namespace
- You want a basic cluster with the Ceph dashboard for monitoring

## Key Configuration Choices

- **Same namespace as operator** (`rook-ceph`) -- the cluster must be in the same namespace as the operator
- **Toolbox enabled** (`true`) -- deploys a debug pod with Ceph CLI tools (`ceph status`, `rados`, etc.)
- **Dashboard enabled** (`true`) -- Ceph web dashboard for monitoring cluster health
- **Monitoring disabled** -- Prometheus integration disabled by default; enable if Prometheus is installed
- **Default storage** -- uses all available OSDs discovered by the operator; configure `cluster` and `blockPools` for fine-tuned control

## Placeholders to Replace

No placeholders -- this preset is directly deployable. Ensure the Rook Ceph Operator is running first.

## Related Presets

- **02-production-with-block-pool** -- Explicit block pool and storage class configuration for production
