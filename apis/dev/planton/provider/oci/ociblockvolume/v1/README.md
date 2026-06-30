# OciBlockVolume

## Overview

OciBlockVolume is an Planton component that deploys an OCI Block Volume along with an optional backup policy assignment. It provides a single declarative manifest to manage a block storage device with configurable performance tiers, autotune policies, cross-region replicas, and scheduled backups.

## Purpose

Block Volumes are OCI's primary persistent block storage for compute instances — database files, application data, boot volumes, and shared-storage clustering. This component wraps the volume and its backup policy assignment into a single resource so that performance tuning, replication, and backup scheduling are configured at deploy time rather than as separate manual steps.

## Key Features

- **Single-resource deployment** — one manifest creates the volume and optionally assigns a backup policy.
- **Performance tiers** — configurable VPUs/GB from Lower Cost (0) through Ultra High Performance (30-120).
- **Autotune policies** — automatic performance adjustment based on detachment state or workload demand.
- **Cross-region replicas** — asynchronous replication to availability domains in other regions for disaster recovery.
- **Backup policy assignment** — link to Oracle-defined (Gold, Silver, Bronze) or custom backup policies for scheduled backups.
- **Customer-managed encryption** — optional KMS keys for the volume, cross-region replicas, and cross-region backups.
- **SCSI persistent reservations** — support for shared-storage clustering scenarios such as Oracle RAC.
- **Foreign key references** — `compartmentId`, `kmsKeyId`, `backupPolicyId`, `xrcKmsKeyId`, and replica `xrrKmsKeyId` support `valueFrom` to reference other Planton-managed resources.

## Constraints

- `availabilityDomain` is immutable — changing it forces recreation.
- `sizeInGbs` must be at least 50 (the OCI minimum).
- `vpusPerGb` must be 0, 10, 20, or 30-120 in increments of 10.
- `xrcKmsKeyId` changes force recreation.
- `maxVpusPerGb` must be > 0 when `autotuneType` is `performance_based`.
- The volume and any attached compute instance must reside in the same availability domain.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Development data volume | Minimal 50 GB volume with Balanced tier |
| High-IOPS database storage | Higher Performance (20 VPUs/GB) or Ultra High Performance (30-120) with KMS encryption |
| Cost-optimized detachable storage | Detached-volume autotune policy reduces VPUs to 0 when not in use |
| Disaster recovery | Cross-region replicas with encrypted replication |
| Compliance backups | Gold/Silver/Bronze backup policy with optional cross-region copy |
| Shared-storage clustering | SCSI persistent reservations for Oracle RAC |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels`, including `resource_kind`, `resource_id`, `organization`, and `environment`.
- **KMS encryption** — customer-managed keys for encryption at rest when regulatory requirements exceed Oracle-managed defaults.
- **Cross-region replication** — asynchronous replicas for RPO-based disaster recovery across OCI regions.
- **Backup scheduling** — backup policy assignment automates daily/weekly/monthly backups per Oracle-defined or custom policies.
