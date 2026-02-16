# AWS FSx for OpenZFS File System Resource Kind

**Date**: February 16, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Added the AwsFsxOpenzfsFileSystem resource kind (enum 292, id_prefix `awsfxz`) to OpenMCF, enabling fully managed NFS file system deployments built on the OpenZFS file system. The component supports SINGLE_AZ_1, SINGLE_AZ_2, and MULTI_AZ_1 deployment types with configurable NFS exports, ZSTD/LZ4 compression, per-user/group quotas, provisioned IOPS, and automatic backups.

## Problem Statement / Motivation

FSx for OpenZFS is AWS's general-purpose NFS file system service, positioned between EFS (simpler, serverless) and FSx for Lustre (HPC-optimized). It fills a gap in OpenMCF's AWS storage coverage for workloads needing standard NFS with advanced ZFS features like snapshots, cloning, and compression.

### Pain Points

- No OpenMCF component for deploying managed NFS with ZFS features
- Teams needing NFS storage with compression, quotas, or Multi-AZ HA had no declarative option
- FSx OpenZFS is a fundamentally different service from FSx Lustre — separate Terraform resource, distinct schema, unique sub-resource hierarchy

## Solution / What's New

A complete deployment component following the FSx family pattern established by AwsFsxLustreFileSystem (R29a). The OpenZFS component differs from Lustre in several key ways:

- **Multi-AZ support**: MULTI_AZ_1 deployment with automatic failover, preferred subnet, route table management, and floating IP
- **NFS exports**: Client-level access control with IP/CIDR/wildcard and mount options
- **ZFS features**: ZSTD/LZ4 compression, configurable record size (4-1024 KiB), per-user/group quotas
- **Throughput model**: Absolute MB/s (not per-TiB like Lustre)

## Implementation Details

### Proto API (4 files, 6 nested messages, 10 CEL validations)

- `spec.proto` — 17 fields across file system core, networking, encryption, IOPS, root volume configuration, backup, and maintenance
- Nested messages: `DiskIopsConfiguration`, `RootVolumeConfiguration`, `NfsExports`, `NfsClientConfiguration`, `UserAndGroupQuota`
- CEL cross-field validations: deployment type enum, Multi-AZ field restrictions (preferred_subnet, route_tables, endpoint_ip_range), IOPS mode validation, compression type, record size, quota type

### Pulumi Module (4 Go files)

Single `fsx.OpenZfsFileSystem` resource with comprehensive field mapping including nested blocks for NFS exports, quotas, and IOPS configuration. Follows the Lustre sibling pattern.

### Terraform Module (5 HCL files)

Feature parity with Pulumi module. Dynamic blocks for `disk_iops_configuration`, `root_volume_configuration` with nested `nfs_exports` and `user_and_group_quotas`.

### Validation Tests

52 tests covering all CEL rules, field-level constraints, happy paths for all deployment types, nested message validations, and edge cases.

### Key Design Decisions

- **INTELLIGENT_TIERING excluded** from v1 (MULTI_AZ_1 only, no explicit storage capacity, requires read_cache_configuration — too new and niche)
- **Child volumes NOT bundled** — independent lifecycle, root volume configured inline
- **subnet_ids is repeated** (not singular like Lustre) to support MULTI_AZ_1's two-subnet requirement
- **Default deployment type**: SINGLE_AZ_2 (latest generation, recommended for new workloads)

## Benefits

- Declarative NFS file system management with compression, quotas, and Multi-AZ HA
- Cross-resource references via StringValueOrRef for subnets, security groups, KMS keys
- 3 presets covering development, production, and HA use cases
- Production documentation with 5 examples including valueFrom patterns

## Impact

- **New resource kind**: AwsFsxOpenzfsFileSystem registered as enum 292
- **39 files, 4,244 lines** added to OpenMCF
- **52 validation tests** all passing
- Continues the FSx family (R29a Lustre done, R29b OpenZFS done, R29c-f pending)

## Related Work

- AwsFsxLustreFileSystem (R29a) — sibling FSx component for HPC workloads
- AwsElasticFileSystem (R11) — simpler serverless NFS alternative
- Part of the AWS resource expansion project (20260215.02.sp.aws-resource-expansion)

---

**Status**: Production Ready
