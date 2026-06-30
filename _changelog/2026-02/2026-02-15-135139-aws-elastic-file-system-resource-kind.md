# AWS Elastic File System Resource Kind

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi Module, Terraform Module, Documentation

## Summary

Added AwsElasticFileSystem as a new AWS resource kind in Planton, providing managed NFS file storage that bundles the file system with mount targets, access points, backup policy, and resource policy into a single deployable component. This is the thirteenth new AWS resource kind in the expansion project (R11).

## Problem Statement / Motivation

AWS Elastic File System (EFS) is the standard solution for shared persistent storage across EKS pods, ECS tasks, Lambda functions, and EC2 instances. Without an EFS component in Planton, users had no declarative way to provision shared file storage that multiple compute resources could mount simultaneously.

### Pain Points

- No shared storage component for Kubernetes workloads (EFS CSI driver requires `file_system_id`)
- Lambda functions needing persistent file access had no EFS provisioning path
- ECS tasks sharing data between containers lacked a declarative EFS option
- Mount target creation (one per AZ) was manual and error-prone

## Solution / What's New

A complete AwsElasticFileSystem deployment component that bundles 5 AWS resources into a single declarative manifest:

- `aws_efs_file_system` — the file system with encryption, throughput mode, lifecycle policies
- `aws_efs_mount_target` — one per subnet, auto-created from `subnet_ids`
- `aws_efs_access_point` — optional, per-application POSIX identity and root directory isolation
- `aws_efs_backup_policy` — automatic daily backups via AWS Backup
- `aws_efs_file_system_policy` — IAM resource policy (e.g., enforce encryption in transit)

## Implementation Details

### Proto API (4 files)

- `spec.proto`: 14 top-level fields, 4 nested messages, 10 CEL validations
- `stack_outputs.proto`: 8 outputs including 4 map outputs (mount target and access point data keyed by subnet_id and name)
- `api.proto`: KRM wiring (apiVersion, kind, metadata, spec, status)
- `stack_input.proto`: Target + AWS provider config

### Key Spec Fields

| Category | Fields |
|----------|--------|
| Core | encrypted, kms_key_id, performance_mode, throughput_mode, provisioned_throughput_in_mibps, availability_zone_name |
| Lifecycle | transition_to_ia, transition_to_archive, transition_to_primary_storage_class |
| Networking | subnet_ids (required), security_group_ids |
| Access Points | name, posix_user (uid/gid), root_directory (path + creation_info) |
| Policy | backup_enabled, policy (google.protobuf.Struct) |

### CEL Validations (10)

- performance_mode and throughput_mode in-list validation
- Provisioned throughput bidirectional dependency (mode requires value, value requires mode)
- KMS key requires encryption enabled
- Archive transition requires IA transition (AWS constraint: files must pass through IA)
- Lifecycle value set validation
- Primary storage class restricted to AFTER_1_ACCESS

### Pulumi Module (7 Go files)

- `main.go`, `locals.go`, `outputs.go`, `file_system.go`, `mount_target.go`, `access_point.go`, `policy.go`
- Mount targets iterate `subnet_ids`, creating one `efs.NewMountTarget` per subnet
- Access points iterate the repeated message with POSIX user and root directory config
- Policy serializes `google.protobuf.Struct` to JSON via `json.Marshal`

### Terraform Module (5 files)

- `for_each` on subnet_ids for mount targets
- `for_each` on access_point_map (keyed by name) for access points
- Dynamic lifecycle_policy blocks
- Conditional backup and file system policy via count

### Validation Tests

22 spec tests (12 happy path + 10 failure scenarios), all passing via Ginkgo + protovalidate.

## Benefits

- **Shared storage for Kubernetes**: EKS pods can mount EFS via the CSI driver using `file_system_id`
- **Lambda file access**: Access point ARNs enable Lambda file system configurations
- **Cost optimization**: Lifecycle policies auto-tier files to IA (92% savings) and Archive (96% savings)
- **Security**: Resource policies enforce encryption in transit; POSIX access points enforce least-privilege

## Impact

- **New resource kind**: AwsElasticFileSystem (enum 290)
- **Files**: ~35 source files (proto, Go, Terraform, docs, presets)
- **Tests**: 22 validation tests
- **Downstream enablement**: Unlocks EFS-backed storage for EKS, ECS, and Lambda infra charts

## Related Work

- AwsElasticIp (R10) — previous resource in the queue
- AwsEipAssociation (R10a) — skipped (EIP binding handled inline by consumer resources)
- AwsVpc — provides subnets referenced by mount targets
- AwsSecurityGroup — controls NFS traffic to mount targets
- AwsKmsKey — optional custom encryption key

---

**Status**: ✅ Production Ready
