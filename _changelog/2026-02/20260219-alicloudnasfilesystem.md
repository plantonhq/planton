# AlicloudNasFileSystem Component Added

**Date**: 2026-02-19
**Component**: AlicloudNasFileSystem
**Enum**: 3051
**ID Prefix**: acnas

## Summary

Added the AlicloudNasFileSystem deployment component -- the second Storage-tier resource in the Alibaba Cloud catalog. This component manages an Alibaba Cloud NAS file system with a VPC mount target and optional custom access group with IP-based access rules. It supports both standard (auto-scaling) and extreme (dedicated throughput) file system types, NFS and SMB protocols, and optional encryption at rest.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudnasfilesystem/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AlicloudNasFileSystem = 3051` in `CloudResourceKind` enum under the Storage category
- 2 nested messages: `AlicloudNasEncryption`, `AlicloudNasAccessRule`

### IaC Modules
- **Pulumi** (Go): Creates NAS file system, conditionally creates access group with access rules when `accessRules` is non-empty, always creates VPC mount target. Handles extreme NAS (sets VPC/VSwitch on file system) vs standard NAS (VPC/VSwitch only on mount target).
- **Terraform** (HCL): `main.tf` for file system + mount target, `access_group.tf` for conditional access group + rules using `count` and `for_each`. Variables, outputs, locals, and provider in separate files.

### Tests
- Ginkgo/Gomega spec validation tests: 21 specs covering valid inputs (minimal NFS, full config with encryption and access rules, extreme NAS, SMB protocol, Premium storage, all_squash access rule), invalid inputs (missing region, missing protocol_type, invalid protocol_type, invalid storage_type, invalid file_system_type, missing vpc_id, missing vswitch_id, invalid encrypt_type, missing access_rule source_cidr_ip, invalid rw_access_type, invalid user_access_type, wrong api_version, wrong kind, missing metadata, missing spec)

### Documentation
- README.md with configuration reference, file system types table, access rule fields, and related components
- examples.md with 4 YAML examples (minimal NFS, production encrypted with access rules, extreme NAS for HPC, SMB Capacity for archival)
- catalog-page.md with full configuration reference and examples
- docs/README.md with comprehensive research documentation covering all 14 NAS provider resources, storage types, encryption model, access control model, and mount target behavior
- 2 presets: standard-nfs, production-encrypted

## Spec Design Decisions

- **`file_system_type` added (not in T02)**: The provider distinguishes between standard and extreme NAS with fundamentally different behaviors (auto-scaling vs fixed capacity, different storage type values). Defaulting to "standard" covers the 80% case while allowing extreme NAS for high-throughput workloads.
- **`access_rules` replaces T02's `access_group_name`**: Exposing the access group name leaks a provider implementation detail. Instead, when access_rules are specified, a custom access group is auto-created. When omitted, the default VPC group provides full RDWR access from all VPC IPs. This is cleaner UX.
- **Encryption included (not in T02)**: Encryption is production-critical. Modeled as an optional `AlicloudNasEncryption` message (same pattern as OssBucket), supporting NAS-managed (1) and KMS customer-managed (2) encryption.
- **`capacity` and `zone_id` added (not in T02)**: Required for extreme NAS but optional/ignored for standard NAS. Standard NAS auto-assigns zones and auto-scales capacity.
- **Storage type values expanded**: T02 listed only "Performance" and "Capacity". The provider supports "Premium" for standard NAS and "standard"/"advance" for extreme NAS.
- **CPFS excluded**: Cloud Parallel File System is a niche HPC product with distinct requirements. Can be added as a separate component if needed.
- **Composite bundling (DD07)**: File system + access group + access rules + mount target are bundled because a file system without a mount target is unreachable.

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (21/21 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
