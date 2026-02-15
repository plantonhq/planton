---
title: "Production Ceph Cluster with Block Pool"
description: "This preset creates a production Ceph cluster with an explicit replicated block pool, a default Kubernetes StorageClass, and Prometheus monitoring enabled."
type: "preset"
rank: "02"
presetSlug: "02-production-with-block-pool"
componentSlug: "rook-ceph-cluster"
componentTitle: "Rook Ceph Cluster"
provider: "kubernetes"
icon: "package"
order: 2
---

# Production Ceph Cluster with Block Pool

This preset creates a production Ceph cluster with an explicit replicated block pool, a default Kubernetes StorageClass, and Prometheus monitoring enabled.

## When to Use

- Production clusters needing Ceph block storage (RBD) for PersistentVolumeClaims
- Environments where a named StorageClass is required for workloads to request Ceph storage
- Clusters with Prometheus for Ceph health monitoring

## Key Configuration Choices

- **Replicated block pool** (3 replicas) -- data is replicated across 3 OSDs for durability; tolerates 1 OSD failure
- **Default StorageClass** (`ceph-block`, `isDefault: true`) -- PVCs without a storageClassName will use this class
- **Volume expansion enabled** -- PVCs can be resized after creation
- **Monitoring enabled** -- Prometheus ServiceMonitor for Ceph metrics
- **Toolbox and dashboard** -- operational tools for cluster management

## Placeholders to Replace

No placeholders -- this preset is directly deployable. Ensure the Rook Ceph Operator is running and at least 3 OSD nodes are available.

## Related Presets

- **01-standard** -- Minimal cluster without explicit block pool configuration
