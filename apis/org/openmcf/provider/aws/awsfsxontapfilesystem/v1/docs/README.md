# AWS FSx for ONTAP: Architecture Reference

This document provides a deep technical reference for Amazon FSx for NetApp ONTAP as deployed via the AwsFsxOntapFileSystem API. It covers the file system hierarchy, deployment types, HA pairs, storage options, networking, endpoints, backup strategy, encryption, and cost optimization.

---

## 1. Architecture: File System → SVMs → Volumes

### Hierarchy

FSx for ONTAP follows a three-tier model:

```
File System (this component)
  └── Storage Virtual Machine (SVM) — AwsFsxOntapStorageVirtualMachine
        └── Volume — AwsFsxOntapVolume
```

- **File system:** The top-level ONTAP cluster. Defines deployment type, capacity, throughput, networking, and encryption. One file system can host multiple SVMs.
- **Storage Virtual Machine (SVM):** A logical storage server. Exposes NFS, SMB, and/or iSCSI protocols. Each SVM has its own namespace, LIFs (Logical Interfaces), and volumes.
- **Volume:** A logical storage container within an SVM. Volumes are what clients mount (NFS export path, SMB share, or iSCSI LUN).

**Key design note:** This OpenMCF component manages only the file system. SVMs and volumes are separate resources with their own lifecycle. Create them after the file system is provisioned, referencing `status.outputs.file_system_id`.

### Data Flow

```
Client (EC2, ECS, VMware)
  → NFS/SMB/iSCSI
  → LIF (Logical Interface) on SVM
  → Volume
  → Aggregate (ONTAP storage pool)
  → SSD/HDD storage
```

---

## 2. Deployment Types

### SINGLE_AZ_1

- **Availability:** Single AZ. No cross-AZ failover.
- **HA pairs:** 1–12. Adding HA pairs requires replacement.
- **Subnets:** Exactly one.
- **Status:** First generation. Prefer SINGLE_AZ_2 for new deployments.

### SINGLE_AZ_2

- **Availability:** Single AZ. Latest generation.
- **HA pairs:** 1–12. Can increase HA pairs **without replacement** (in-place scale-out).
- **Subnets:** Exactly one.
- **Use case:** Most workloads. Recommended default for single-AZ.

### MULTI_AZ_1

- **Availability:** Multi-AZ with automatic failover across two AZs.
- **HA pairs:** Fixed at 1 (standby in second AZ).
- **Subnets:** Exactly two, in different AZs.
- **Requirements:** `preferred_subnet_id`, `endpoint_ip_address_range`.
- **Status:** First generation. Prefer MULTI_AZ_2 for new multi-AZ deployments.

### MULTI_AZ_2

- **Availability:** Multi-AZ with automatic failover. Latest generation.
- **HA pairs:** Fixed at 1.
- **Subnets:** Exactly two, in different AZs.
- **Requirements:** `preferred_subnet_id`, `endpoint_ip_address_range`.
- **Use case:** Mission-critical workloads requiring high availability.

### Decision Matrix

| Criteria | SINGLE_AZ_1 | SINGLE_AZ_2 | MULTI_AZ_1 | MULTI_AZ_2 |
|----------|-------------|-------------|-----------|-----------|
| Generation | First | Latest | First | Latest |
| Cross-AZ failover | No | No | Yes | Yes |
| HA pairs | 1–12 (replace to add) | 1–12 (in-place add) | 1 | 1 |
| Subnets | 1 | 1 | 2 | 2 |
| Scale-out | Replace to add HA pairs | In-place add HA pairs | N/A | N/A |

---

## 3. HA Pairs Explained

### What Is an HA Pair?

An HA (High Availability) pair is a pair of file servers (nodes) that provide redundancy within a single AZ. Each pair contributes independent throughput and IOPS capacity.

### Scale-Out (Single-AZ Only)

- **SINGLE_AZ_1 / SINGLE_AZ_2:** 1–12 HA pairs.
- **Total throughput** = `throughput_capacity_per_ha_pair` × `ha_pairs`.
- **Example:** 4 HA pairs × 512 MB/s = 2048 MB/s aggregate throughput.
- **SINGLE_AZ_2:** Can increase `ha_pairs` without replacing the file system. SINGLE_AZ_1 requires replacement.

### Multi-AZ Fixed at 1 HA Pair

- **MULTI_AZ_1 / MULTI_AZ_2:** Always 1 HA pair.
- The pair spans two AZs: active node in preferred subnet, standby in the other.
- Failover is automatic; no scale-out via HA pairs.

### Throughput per HA Pair

Valid values: 128, 256, 384, 512, 768, 1024, 1536, 2048, 3072, 4096, 6144 MB/s.

Higher tiers cost more. Choose based on workload I/O profile. Can be changed after creation for SINGLE_AZ_2 and MULTI_AZ_2.

---

## 4. Storage Types

### SSD

- **Latency:** Sub-millisecond.
- **Capacity range:** 1024–1,048,576 GiB (1 TiB – 1 PiB).
- **Use case:** Performance-sensitive workloads (databases, VMware, active data).
- **IOPS:** AUTOMATIC (3 IOPS/GiB) or USER_PROVISIONED (up to 2,400,000).

### HDD with Intelligent Tiering

- **Latency:** Higher (single-digit milliseconds).
- **Capacity range:** 1024–1,048,576 GiB.
- **Use case:** Throughput-oriented workloads (data lakes, backups, infrequently accessed data).
- **Intelligent tiering:** ONTAP automatically caches hot data on SSD.

**ForceNew:** Storage type cannot be changed after creation.

### Data Reduction

ONTAP provides built-in compression and deduplication. Typical effective capacity is 2–5× the provisioned capacity for many workloads.

---

## 5. Networking

### Single-AZ

- **Subnets:** Exactly one subnet.
- **ENIs:** One per HA pair (e.g., 4 HA pairs = 4 ENIs).
- **Placement:** All nodes in the same AZ.

### Multi-AZ

- **Subnets:** Exactly two subnets in different AZs.
- **ENIs:** Two (one per AZ).
- **Endpoint IP address range:** CIDR block within the VPC for floating IPs. Must not overlap with existing subnets. AWS assigns IPs from this range for seamless failover.
- **Route tables:** Optional. `route_table_ids` specify route tables that need routes to the file system. AWS manages routes for failover.

### Security Groups

Must allow:

| Port | Protocol | Purpose |
|------|----------|---------|
| 111 | TCP | Portmapper (NFS) |
| 635 | TCP | mountd (NFS) |
| 2049 | TCP | NFS |
| 4045-4046 | TCP | NFS lock/status |
| 445 | TCP | SMB |
| 3260 | TCP | iSCSI |
| 443 | TCP | ONTAP REST API |

---

## 6. Endpoints

### Management Endpoint

- **Purpose:** ONTAP CLI (SSH) and REST API access.
- **Outputs:** `management_dns_name`, `management_ip_addresses`.
- **Access:** `ssh fsxadmin@<management_dns_name>` (requires `fsx_admin_password` in spec).
- **Use cases:** LIF management, SnapMirror configuration, aggregate monitoring, advanced administration.

### Intercluster Endpoint

- **Purpose:** NetApp SnapMirror replication between FSx for ONTAP file systems (same or cross-region).
- **Outputs:** `intercluster_dns_name`, `intercluster_ip_addresses`.
- **Use case:** Hybrid cloud replication to/from on-premises NetApp, or disaster recovery between AWS regions.

### Data Endpoints

Data access (NFS, SMB, iSCSI) is provided by **SVMs**, not the file system directly. Create an SVM and volumes to expose data endpoints.

---

## 7. ONTAP CLI and REST API Access

### Enabling Access

Set `fsx_admin_password` in the spec (8–50 characters). This enables:

- **SSH:** `ssh fsxadmin@<management_dns_name>`
- **REST API:** `https://<management_dns_name>/api`

### Security

- The password is sensitive and is not returned in read operations.
- Omit `fsx_admin_password` if ONTAP CLI access is not needed (reduces attack surface).

### Common CLI Operations

- **System health:** `system health status show`
- **Aggregate info:** `storage aggregate show`
- **SVM list:** `vserver show`
- **Volume list:** `volume show`
- **SnapMirror:** Configure via CLI or REST API for replication.

---

## 8. Backup Strategy

### FSx Automatic Backups

- **Retention:** `automatic_backup_retention_days` (0–90). Set 0 to disable.
- **Window:** `daily_automatic_backup_start_time` in `HH:MM` UTC.
- **Mechanics:** Incremental. Stored in AWS-managed location. Restore creates a new file system.

### ONTAP Snapshots

- **Independent:** ONTAP's built-in snapshots are separate from FSx backups.
- **Configured on volumes:** Create snapshots via SVM/volume configuration or ONTAP CLI.
- **Use case:** Point-in-time recovery, cloning, SnapMirror source.

### Recommendation

- **Development:** Disable FSx backups (`automatic_backup_retention_days: 0`). Use ONTAP snapshots if needed.
- **Production:** Enable FSx backups (7–30 days) plus ONTAP snapshots for layered protection.

---

## 9. Encryption at Rest

- **Default:** All FSx for ONTAP file systems are encrypted at rest with an AWS-managed key.
- **Customer-managed key:** Set `kms_key_id` to use a customer-managed KMS key. **ForceNew** — cannot be changed after creation.
- **Key requirements:** The KMS key must be in the same region as the file system and allow FSx to use it.

---

## 10. Cost Optimization Tips

1. **Right-size throughput:** Start with the lowest tier that meets your needs. Throughput can be scaled up after creation for SINGLE_AZ_2 and MULTI_AZ_2.
2. **Use HDD for cold data:** Large archives and infrequently accessed data benefit from HDD's lower cost per GiB.
3. **Leverage data reduction:** ONTAP compression and deduplication reduce effective storage cost (2–5× typical).
4. **Manage backup retention:** Set `automatic_backup_retention_days` to the minimum required by recovery objectives.
5. **Single-AZ for non-critical workloads:** MULTI_AZ adds cost for the standby server and cross-AZ replication.
6. **Scale-out vs. scale-up:** For single-AZ, adding HA pairs increases throughput. Ensure workload can utilize the additional capacity before scaling.
7. **AUTOMATIC IOPS:** Use `disk_iops_configuration.mode: AUTOMATIC` unless you have a demonstrated need for USER_PROVISIONED IOPS.

---

## 11. Summary

| Topic | Key Takeaway |
|-------|--------------|
| **Hierarchy** | File system → SVM → Volume. This component manages file system only. |
| **Deployment types** | SINGLE_AZ_2 for most; MULTI_AZ_2 for HA. |
| **HA pairs** | Single-AZ: 1–12 (scale-out). Multi-AZ: fixed at 1. |
| **Throughput** | Total = per-HA-pair × ha_pairs. Valid tiers: 128–6144 MB/s. |
| **Storage** | SSD (performance) or HDD (cost). 1024–1,048,576 GiB. |
| **Networking** | Single-AZ: 1 subnet. Multi-AZ: 2 subnets + endpoint_ip_address_range. |
| **Endpoints** | Management (CLI/API), intercluster (SnapMirror). Data via SVMs. |
| **Backups** | FSx backups (0–90 days) and ONTAP snapshots (on volumes). |
| **Encryption** | Always on. Optional customer-managed KMS. |

For API reference and examples, see the parent [README.md](../README.md) and [examples.md](../examples.md).
