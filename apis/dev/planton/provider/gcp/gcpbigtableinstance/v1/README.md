# GcpBigtableInstance

Planton component for provisioning Google Cloud Bigtable instances with one or more clusters.

## Overview

Cloud Bigtable is Google Cloud's fully managed, wide-column NoSQL database designed for large analytical and operational workloads. It provides consistent sub-10ms latency, scales to billions of rows and thousands of columns, and is ideal for time-series data, IoT, ad-tech, fintech, and machine-learning feature stores.

This component bundles a Bigtable instance (the logical container for data) with one or more clusters (the physical replicas serving the data). An instance without at least one cluster cannot store or serve data, so they are provisioned together as a single unit.

## Key Features

- **Instance + clusters** — One component provisions the instance and all its clusters; no instance without compute
- **Multi-cluster replication** — Up to 8 clusters across zones and regions for automatic replication and failover
- **Flexible scaling** — Fixed node count, autoscaling (CPU/storage targets), or automatic allocation per cluster
- **Storage types** — SSD (low latency, default) or HDD (lower cost, batch analytics)
- **CMEK encryption** — Customer-managed encryption keys per cluster
- **Node scaling factor** — 1X (single-node increments) or 2X (two-node increments) per cluster
- **Deletion protection** — Enabled by default; prevents accidental destruction of production instances

## Key Fields

| Field | Required | Description |
|-------|----------|-------------|
| `projectId` | Yes | GCP project ID |
| `instanceName` | Yes | Instance name (lowercase, letters, numbers, hyphens; 6–33 chars) |
| `displayName` | No | Human-readable display name; defaults to `instanceName` |
| `deletionProtection` | No | Prevent accidental deletion; defaults to `true` |
| `forceDestroy` | No | Delete all backups when destroying the instance |
| `clusters` | Yes | One or more cluster configurations (min 1) |

## Cluster Options

Each cluster in the `clusters` array supports the following:

| Field | Required | Description |
|-------|----------|-------------|
| `clusterId` | Yes | Unique cluster ID within the instance (6–30 chars) |
| `zone` | Yes | GCP zone (e.g., `us-central1-a`) |
| `numNodes` | No | Fixed node count; mutually exclusive with `autoscalingConfig` |
| `storageType` | No | `SSD` (default) or `HDD` |
| `kmsKeyName` | No | Cloud KMS key for CMEK encryption |
| `nodeScalingFactor` | No | `NodeScalingFactor1X` (default) or `NodeScalingFactor2X` |
| `autoscalingConfig` | No | Autoscaling with `minNodes`, `maxNodes`, `cpuTarget`, `storageTarget` |

## Scaling Modes

Each cluster supports three scaling approaches (mutually exclusive):

1. **Fixed** — Set `numNodes` to a specific value (e.g., 3 nodes)
2. **Autoscaling** — Configure `autoscalingConfig` with CPU/storage targets and min/max node bounds
3. **Automatic** — Omit both; Bigtable auto-allocates based on data footprint

## Storage Types

| Type | Latency | Cost | Use Case |
|------|---------|------|----------|
| SSD | Sub-10ms | Higher | Real-time serving, most workloads |
| HDD | Higher | Lower | Batch analytics, large cold datasets |

Storage type is immutable after cluster creation.

## CMEK Encryption

Each cluster can be encrypted with a Cloud KMS key via `kmsKeyName`. The KMS key region must match the cluster zone's region. CMEK is immutable after creation.

## Immutable Fields

The following cannot be changed after creation; changing them requires recreating the resource:

- `instanceName`
- Each cluster's `zone`, `storageType`, `kmsKeyName`, `nodeScalingFactor`

## Deletion Protection

`deletionProtection` defaults to `true`. Set to `false` before destroying an instance that contains data.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `instance_id` | Fully qualified instance resource name (`projects/{project}/instances/{instance}`) |
| `instance_name` | Short instance name (same as `instanceName` input) |

## Examples

See [examples.md](examples.md) for copy-paste ready YAML manifests.

## Further Reading

- [Research & design documentation](docs/README.md) — Architecture, replication, scaling, CMEK, best practices
