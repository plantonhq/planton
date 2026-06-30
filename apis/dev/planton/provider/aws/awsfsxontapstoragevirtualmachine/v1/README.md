# AwsFsxOntapStorageVirtualMachine

An Planton component that provisions an **Amazon FSx for NetApp ONTAP Storage Virtual Machine (SVM)** — a logical data server within an FSx ONTAP file system providing multi-protocol data access (NFS, SMB, iSCSI).

## What Is an ONTAP SVM?

In the FSx for ONTAP architecture, SVMs are the data access layer:

- **File System** → physical infrastructure (storage, throughput, networking, HA)
- **SVM** → logical data server (protocol endpoints, AD integration, security style)
- **Volume** → data container (capacity, tiering, snapshots, SnapLock)

Each SVM has its own set of network endpoints (NFS, SMB, iSCSI, management), enabling multi-tenancy on a single file system. You can create multiple SVMs on one file system to isolate workloads.

## When to Use

- **NFS file shares** for Linux workloads, Kubernetes PVs, or data pipelines
- **SMB file shares** for Windows workloads, home directories, or .NET applications
- **iSCSI block storage** for databases or applications requiring block-level access
- **Multi-tenancy** — isolate different teams or applications on a shared file system
- **Multi-protocol access** — serve the same data via NFS and SMB simultaneously

## Prerequisites

1. An existing **AwsFsxOntapFileSystem** — the SVM's parent file system
2. Network connectivity between your compute resources and the file system's subnets
3. (For SMB) A self-managed **Active Directory** domain reachable from the VPC

## Quick Start

### Minimal NFS-only SVM

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: my-nfs-svm
  id: awsfxosvm-my-nfs-svm
  org: my-org
  env: dev
spec:
  file_system_id:
    value: fs-0123456789abcdef0
  name: svm_nfs
  root_volume_security_style: UNIX
```

### SMB SVM with Active Directory

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: my-smb-svm
  id: awsfxosvm-my-smb-svm
  org: my-org
  env: prod
spec:
  file_system_id:
    value: fs-0123456789abcdef0
  name: svm_smb
  root_volume_security_style: NTFS
  svm_admin_password: VsAdmin2024!
  active_directory_configuration:
    netbios_name: SVMSMB
    domain_name: corp.example.com
    dns_ips:
      - "10.0.0.1"
      - "10.0.0.2"
    username: svc_fsx_join
    password: ADJoinP@ssw0rd!
```

### Cross-Resource Reference (valueFrom)

```yaml
spec:
  file_system_id:
    valueFrom:
      kind: AwsFsxOntapFileSystem
      name: my-ontap-fs
      fieldPath: status.outputs.file_system_id
```

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `file_system_id` | StringValueOrRef | Yes | — | Parent FSx ONTAP file system ID |
| `name` | string | Yes | — | ONTAP SVM name (1-47 alphanumeric + underscore) |
| `root_volume_security_style` | string | No | `UNIX` | `UNIX`, `NTFS`, or `MIXED` (ForceNew) |
| `svm_admin_password` | string | No | — | vsadmin password (8-50 chars, sensitive) |
| `active_directory_configuration` | object | No | — | AD config for SMB access |

### Active Directory Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `netbios_name` | string | No | auto | NetBIOS name (1-15 chars) |
| `domain_name` | string | Yes | — | AD domain FQDN |
| `dns_ips` | list | Yes | — | AD DNS server IPs (1-3) |
| `username` | string | Yes | — | AD service account username |
| `password` | string | Yes | — | AD service account password (sensitive) |
| `file_system_administrators_group` | string | No | `Domain Admins` | AD group for admin privileges |
| `organizational_unit_distinguished_name` | string | No | Computers | OU DN for computer object |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `svm_id` | SVM identifier |
| `arn` | SVM ARN |
| `uuid` | ONTAP UUID |
| `subtype` | SVM subtype |
| `iscsi_dns_name` | iSCSI endpoint DNS |
| `iscsi_ip_addresses` | iSCSI endpoint IPs |
| `management_dns_name` | Management endpoint DNS |
| `management_ip_addresses` | Management endpoint IPs |
| `nfs_dns_name` | NFS endpoint DNS |
| `nfs_ip_addresses` | NFS endpoint IPs |
| `smb_dns_name` | SMB endpoint DNS (AD only) |
| `smb_ip_addresses` | SMB endpoint IPs (AD only) |

## Security Style Guide

| Style | Use Case | Permissions Model |
|-------|----------|-------------------|
| `UNIX` | Linux/NFS workloads | UNIX mode bits (uid/gid) |
| `NTFS` | Windows/SMB workloads | Windows ACLs (identity-based) |
| `MIXED` | Dual-protocol (NFS+SMB) | Last protocol to set permissions wins |

## Presets

- **01-nfs-unix** — Basic NFS-only SVM (most common)
- **02-smb-windows** — Windows SMB with Active Directory
- **03-multiprotocol** — Dual NFS+SMB with MIXED security style
