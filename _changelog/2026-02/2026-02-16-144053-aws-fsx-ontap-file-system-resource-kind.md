# AWS FSx for ONTAP File System Resource Kind

**Date**: February 16, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Added the AwsFsxOntapFileSystem resource kind (enum 294, id_prefix `awsfxo`) to OpenMCF, enabling declarative provisioning of Amazon FSx for NetApp ONTAP file systems. This is the fourth FSx type in OpenMCF (after Lustre, OpenZFS, Windows), covering the most enterprise-grade variant with multi-protocol access, scale-out HA pairs, and SnapMirror replication support.

## Problem Statement / Motivation

FSx for ONTAP is AWS's most feature-rich managed file system, providing enterprise NAS/SAN capabilities that no other FSx type offers: simultaneous NFS, SMB, and iSCSI protocol access from a single file system, up to 12 HA pairs for petabyte-scale single-AZ deployments, and NetApp's SnapMirror for cross-region replication. Without this component, OpenMCF users needing enterprise storage for VMware Cloud on AWS, database workloads, or hybrid cloud scenarios had no declarative option.

### Pain Points

- No OpenMCF component for ONTAP file systems despite being the most requested FSx type for enterprise workloads
- Users managing ONTAP file systems manually or through raw Terraform without the OpenMCF validation, preset, and cross-reference framework
- The remaining FSx ONTAP sub-resources (SVMs, Volumes) depend on this file system component existing first

## Solution / What's New

A complete deployment component following the OpenMCF ideal state, covering the ONTAP file system resource (the storage/networking fabric). Storage Virtual Machines and Volumes are separate lifecycle resources handled by companion components.

### Key Design Decisions

1. **throughput_capacity_per_ha_pair** instead of the Terraform dual-field approach (`throughput_capacity` vs `throughput_capacity_per_ha_pair` ExactlyOneOf). The per-HA-pair field works universally across all 4 deployment types and is semantically honest about what it represents.

2. **Four deployment types** (unlike 3 in OpenZFS/Windows siblings): SINGLE_AZ_1, SINGLE_AZ_2, MULTI_AZ_1, MULTI_AZ_2. The `_2` variants are latest-generation with better update semantics.

3. **HA pairs** are validated via CEL to only allow > 1 for single-AZ deployment types. Multi-AZ is fixed at 1 HA pair.

4. **No inline volume configuration** -- unlike OpenZFS which has `root_volume_configuration`, ONTAP manages volumes through SVMs (independent lifecycle). This keeps the file system spec focused on the storage fabric.

5. **Backup default 0 days** -- consistent with Lustre and OpenZFS (opt-in approach). ONTAP's built-in snapshots provide point-in-time recovery independently of FSx backups.

## Implementation Details

### Proto API (4 files)

- **spec.proto**: 18 top-level fields + 1 nested message (DiskIopsConfiguration). 10 CEL cross-field validations covering deployment type constraints, HA pair limits, multi-AZ field gating, admin password length, throughput tier validation, and backup time dependencies.
- **stack_outputs.proto**: 10 outputs including ONTAP-specific endpoints (management DNS/IPs for CLI access, intercluster DNS/IPs for SnapMirror replication).
- **api.proto**: KRM envelope with `aws.openmcf.org/v1` API version.
- **stack_input.proto**: Standard stack input with AWS provider config.

### Validation Tests

60 spec tests covering:
- 18 happy path scenarios (all 4 deployment types, HA pairs 1-12, HDD storage, admin password edge cases, full production configs, all valid throughput tiers)
- 12 field-level validation failures (storage bounds, throughput bounds, HA pair bounds, backup retention bounds)
- 30 CEL cross-field validation scenarios (deployment type, storage type, throughput tiers, HA pairs on multi-AZ, preferred subnet gating, endpoint IP gating, route table gating, admin password length, backup time dependencies, IOPS mode constraints)

### Pulumi Module (4 Go files)

- `main.go`: Provider setup, resource creation, 10 exports including nested endpoint extraction via `ApplyT`
- `locals.go`: Tag initialization with CloudResourceKind enum
- `outputs.go`: 10 output key constants
- `file_system.go`: Single `fsx.NewOntapFileSystem` resource with conditional field mapping for all 18 spec fields

### Terraform Module (5 HCL files)

- Feature parity with Pulumi module
- Dynamic `disk_iops_configuration` block
- 10 outputs extracting nested endpoint attributes

### Documentation

- README.md: User-facing with minimal and production YAML examples
- examples.md: 5 examples (minimal, production, scale-out, multi-AZ, cross-resource references)
- docs/README.md: Comprehensive technical reference (architecture, deployment types, HA pairs, storage types, networking, endpoints, backup strategy, cost optimization)
- catalog-page.md: Following exemplar structure

### Presets (3)

1. `01-single-az-development`: 1024 GiB SSD, 128 MB/s, no backups
2. `02-single-az-production`: 2048 GiB SSD, 512 MB/s, 7-day backups
3. `03-multi-az-high-availability`: MULTI_AZ_2, 2048 GiB SSD, 512 MB/s, 7-day backups

## Benefits

- Enterprise storage workloads now have a first-class OpenMCF component with full validation
- Scale-out HA pair support enables petabyte-scale deployments via simple `haPairs` field
- Rich cross-resource references via StringValueOrRef for VPC, security groups, KMS keys
- 10 stack outputs enable downstream SVM and volume components to wire dependencies
- Consistent patterns with OpenZFS/Windows siblings make the FSx family predictable

## Impact

- **Users**: Can now declaratively provision FSx for ONTAP file systems with all 4 deployment types
- **Platform**: Enables the remaining R29e (AwsFsxOntapStorageVirtualMachine) and R29f (AwsFsxOntapVolume) components that depend on this file system
- **Coverage**: AWS provider now has 4 of 6 planned FSx types complete

## Related Work

- AwsFsxLustreFileSystem (R29a) -- completed 2026-02-16
- AwsFsxOpenzfsFileSystem (R29b) -- completed 2026-02-16
- AwsFsxWindowsFileSystem (R29c) -- completed 2026-02-16
- AwsFsxOntapStorageVirtualMachine (R29e) -- next in queue
- AwsFsxOntapVolume (R29f) -- depends on R29e

---

**Status**: Production Ready
**Files**: ~45 files, ~5,500 lines
**Tests**: 60 passing
