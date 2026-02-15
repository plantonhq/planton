# GcpAlloydbCluster

OpenMCF component for provisioning Google Cloud AlloyDB clusters with a bundled primary instance.

## Overview

AlloyDB is Google Cloud's fully managed, PostgreSQL-compatible database designed for demanding enterprise workloads. It delivers high throughput, low latency, and strong consistency while maintaining full PostgreSQL compatibility.

This component bundles an AlloyDB cluster with its primary instance. A cluster without a primary instance cannot serve queries, so they are provisioned together as a single unit.

## Key Features

- **Cluster + primary instance** — One component provisions both; no cluster without compute
- **Networking** — VPC peering via Private Service Access (VPC must have Private Service Access configured)
- **CMEK encryption** — Customer-managed keys at cluster, automated backup, and continuous backup levels
- **Automated backups** — Quantity-based or time-based retention, optional weekly schedule
- **Continuous backup** — Point-in-time recovery (PITR) with configurable recovery window (1–35 days)
- **Primary instance** — `cpu_count` or `machine_type`, ZONAL or REGIONAL availability, query insights, SSL, Auth Proxy enforcement

## Key Fields

| Field | Required | Description |
|-------|----------|-------------|
| `projectId` | Yes | GCP project ID |
| `clusterName` | Yes | Cluster name (lowercase, letters, numbers, hyphens; 2–63 chars) |
| `location` | Yes | GCP region (e.g., `us-central1`) |
| `network` | Yes | VPC network self-link (must have Private Service Access) |
| `databaseVersion` | No | `POSTGRES_14`, `POSTGRES_15`, or `POSTGRES_16`; defaults to latest |
| `primaryInstance` | Yes | Primary instance config (instanceId, cpuCount or machineType, availabilityType) |

## Primary Instance Options

- **Machine sizing** — Use `cpuCount` (2, 4, 8, 16, 32, 64, 96, 128) or `machineType` (e.g., `n2-highmem-4`); mutually exclusive
- **Availability** — `ZONAL` (single zone, lower cost) or `REGIONAL` (multi-zone, automatic failover)
- **Query insights** — Performance monitoring with configurable plan capture and query string length
- **SSL** — `ENCRYPTED_ONLY` (recommended) or `ALLOW_UNENCRYPTED_AND_ENCRYPTED`
- **requireConnectors** — When `true`, enforces AlloyDB Auth Proxy or Language Connectors (IAM-based auth)

## CMEK Encryption

Three independent encryption keys are supported:

1. **Cluster** — `kmsKeyName` for data at rest
2. **Automated backup** — `automatedBackupPolicy.encryptionKmsKeyName` for snapshot backups
3. **Continuous backup** — `continuousBackupConfig.encryptionKmsKeyName` for PITR data

Each can use a different KMS key for compliance and key lifecycle management.

## Immutable Fields

The following cannot be changed after creation; changing them requires recreating the cluster:

- `clusterName`, `location`, `network`, `kmsKeyName`, `primaryInstance.instanceId`

## Deletion Protection

`deletionProtection` defaults to `true`. Set to `false` before destroying a cluster that contains data.

## Examples

See [examples.md](examples.md) for copy-paste ready YAML manifests.

## Further Reading

- [Research & design documentation](docs/README.md) — Architecture, backup strategies, CMEK, best practices
- [Terraform module](iac/tf/README.md) — Terraform implementation details
