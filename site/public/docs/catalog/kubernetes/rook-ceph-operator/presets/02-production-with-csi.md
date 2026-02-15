---
title: "Production Rook Ceph Operator with Explicit CSI Configuration"
description: "This preset deploys the Rook Ceph Operator with all CSI driver options explicitly configured. Use this when you need full control over which storage drivers are enabled and want every setting..."
type: "preset"
rank: "02"
presetSlug: "02-production-with-csi"
componentSlug: "rook-ceph-operator"
componentTitle: "Rook Ceph Operator"
provider: "kubernetes"
icon: "package"
order: 2
---

# Production Rook Ceph Operator with Explicit CSI Configuration

This preset deploys the Rook Ceph Operator with all CSI driver options explicitly configured. Use this when you need full control over which storage drivers are enabled and want every setting documented in your manifest for auditability.

## When to Use

- Production clusters where every operator setting must be explicitly declared
- You need to customize CSI driver behavior (e.g., disable CephFS, enable NFS, or adjust provisioner replicas)
- GitOps workflows where implicit defaults are undesirable

## Key Configuration Choices

- **Operator version** (`v1.16.6`) -- pinned for reproducibility; update deliberately
- **CRDs enabled** (`true`) -- the operator manages its own CRDs; set to `false` only if managing CRDs externally
- **RBD driver** (`true`) -- enables Ceph block storage via CSI, the most common Ceph storage type
- **CephFS driver** (`true`) -- enables Ceph shared filesystem via CSI, useful for ReadWriteMany volumes
- **Host networking** (`true`) -- CSI node plugins use host networking for direct access to Ceph cluster; recommended for performance
- **Provisioner replicas** (`2`) -- HA for the CSI provisioner pods; increase for very large clusters
- **CSI Addons** (`false`) -- additional CSI functionality; enable if you need volume replication or reclaimspace
- **NFS driver** (`false`) -- NFS gateway via CSI; enable only if you specifically need NFS access to Ceph

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults. Adjust CSI driver flags based on your storage requirements.

## Related Presets

- **01-standard** -- Minimal configuration that relies on proto defaults for CSI settings
