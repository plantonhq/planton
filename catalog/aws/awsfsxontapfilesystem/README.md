# AwsFsxOntapFileSystem

A catalog kind that provisions an **Amazon FSx for NetApp ONTAP file system** — enterprise-grade, fully managed shared storage with multi-protocol access (NFS, SMB, iSCSI), instant snapshots, cloning, SnapMirror replication, and built-in compression and deduplication.

## What Is an FSx ONTAP File System?

In the FSx for ONTAP architecture, the file system is the physical foundation:

- **File System** (this kind) → physical infrastructure (storage, throughput, networking, HA)
- **SVM** (`AwsFsxOntapStorageVirtualMachine`) → logical data server (protocol endpoints, AD integration, security style)
- **Volume** (`AwsFsxOntapVolume`) → data container (capacity, tiering, snapshots, SnapLock)

One file system hosts multiple SVMs, each with its own endpoints and volumes — the multi-tenancy unit for teams and applications sharing the same provisioned capacity and throughput.

## When to Use

- **Enterprise shared storage** for NFS, SMB, and iSCSI clients from one system
- **Database storage** (Oracle, SAP, SQL Server) on shared NFS or iSCSI with sub-millisecond latency
- **VMware Cloud on AWS** datastores
- **Hybrid cloud** — SnapMirror replication to and from on-premises NetApp
- **Scale-out throughput** — SINGLE_AZ_2 grows to 12 HA pairs, 1 PiB, and tens of GB/s in one namespace

## Choosing a Shape

| Shape | Deployment type | Why |
|-------|-----------------|-----|
| Default | `SINGLE_AZ_2` | Current generation; in-place HA-pair scale-out (1–12) |
| High availability | `MULTI_AZ_2` | Automatic cross-AZ failover with floating endpoints |
| Legacy estate match | `SINGLE_AZ_1` / `MULTI_AZ_1` | First generation — 192 TiB cap, single HA pair, destructive throughput scaling |

## Prerequisites

1. A subnet (single-AZ) or two subnets in different AZs (multi-AZ) — `AwsSubnet`
2. (Recommended) A security group allowing NFS/SMB/iSCSI/management traffic — `AwsSecurityGroup`
3. (Optional) A customer-managed KMS key — `AwsKmsKey`

## Quick Start

### Minimal single-AZ file system

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxOntapFileSystem
metadata:
  name: my-ontap-fs
  id: awsfxo-my-ontap-fs
  org: my-org
  env: dev
spec:
  region: us-west-2
  storage_capacity_gib: 1024
  throughput_capacity_per_ha_pair: 384
  subnet_ids:
    - value: subnet-0123456789abcdef0
```

### Multi-AZ with automatic failover

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxOntapFileSystem
metadata:
  name: ha-ontap-fs
  id: awsfxo-ha-ontap-fs
  org: my-org
  env: prod
spec:
  region: us-west-2
  deployment_type: MULTI_AZ_2
  storage_capacity_gib: 4096
  throughput_capacity_per_ha_pair: 768
  subnet_ids:
    - value: subnet-az-a
    - value: subnet-az-b
  preferred_subnet_id:
    value: subnet-az-a
  endpoint_ip_address_range: 198.19.255.0/24
  route_table_ids:
    - value: rtb-0123456789abcdef0
```

### Cross-resource references (valueFrom)

```yaml
spec:
  subnet_ids:
    - valueFrom:
        kind: AwsSubnet
        name: private-subnet-a
        fieldPath: status.outputs.subnet_id
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: ontap-sg
        fieldPath: status.outputs.security_group_id
```

## Spec Highlights

| Field | Required | Notes |
|-------|----------|-------|
| `region`, `storage_capacity_gib`, `subnet_ids` | Yes | Capacity: 1024 GiB per HA pair up to 512 TiB per pair (192 TiB gen1 cap) |
| `throughput_capacity` XOR `throughput_capacity_per_ha_pair` | Exactly one | Whole-system (gen1) vs per-pair (gen2 tiers 384/768/1536/3072/6144) |
| `deployment_type` | No (default `SINGLE_AZ_2`) | ForceNew |
| `ha_pairs` | No (default 1) | >1 only on SINGLE_AZ_2; added in place |
| `preferred_subnet_id` | Multi-AZ only | Required there, invalid for single-AZ |
| `fsx_admin_password` | No | Enables the ONTAP CLI/REST API (sensitive) |
| `disk_iops_configuration` | No | AUTOMATIC (3 IOPS/GiB) or USER_PROVISIONED up to 2.4M |

Backup-skip and tag-copy decisions are volume-scoped in ONTAP — configure them on `AwsFsxOntapVolume`.

## Outputs

| Output | Description |
|--------|-------------|
| `file_system_id` | The join key for SVMs and other consumers |
| `file_system_arn` | ARN for IAM resource-level permissions |
| `management_dns_name` / `management_ip_addresses` | ONTAP CLI (SSH) and REST API endpoint |
| `intercluster_dns_name` / `intercluster_ip_addresses` | SnapMirror replication endpoint |
| `network_interface_ids` | ENIs (1 per HA pair single-AZ; 2 multi-AZ) |
| `vpc_id`, `owner_id` | Placement and account identity |

Data-access endpoints (NFS/SMB/iSCSI DNS names) live on the SVM, not the file system.

## Presets

- **01-single-az-development** — smallest/cheapest shape, no backups
- **02-single-az-production** — KMS encryption, backups, admin access
- **03-multi-az-high-availability** — MULTI_AZ_2 with floating endpoints
- **04-single-az-scale-out** — 4 HA pairs, 6 GB/s aggregate, provisioned IOPS

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
