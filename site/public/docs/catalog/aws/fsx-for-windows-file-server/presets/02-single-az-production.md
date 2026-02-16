---
title: "Single-AZ Production FSx Windows"
description: "SINGLE_AZ_2 SSD file system with 500 GiB and 256 MB/s throughput. Configured for production workloads with self-managed Active Directory, audit logging for compliance, customer-managed KMS..."
type: "preset"
rank: "02"
presetSlug: "02-single-az-production"
componentSlug: "fsx-for-windows-file-server"
componentTitle: "FSx for Windows File Server"
provider: "aws"
icon: "package"
order: 2
---

# Single-AZ Production FSx Windows

SINGLE_AZ_2 SSD file system with 500 GiB and 256 MB/s throughput. Configured for production workloads with self-managed Active Directory, audit logging for compliance, customer-managed KMS encryption, and 7-day backup retention.

## When to Use

- Production Windows application file storage (home directories, shared drives)
- .NET applications or SQL Server requiring managed SMB file shares
- Environments with existing on-premises or EC2-hosted Active Directory
- Compliance-sensitive workloads requiring audit trails of file access

## What It Configures

- **SINGLE_AZ_2** — Latest generation single-AZ deployment with full throughput range
- **500 GiB SSD** — Sub-millisecond latency with room for growth (can increase up to 65536 GiB)
- **256 MB/s throughput** — Sufficient for moderate production workloads. Can be scaled up after creation
- **Self-managed AD** — Joins an existing on-premises or EC2-hosted Active Directory domain with direct credentials
- **Customer-managed KMS** — Encryption at rest with your own KMS key for audit trails and key rotation
- **Audit logging** — File access events logged at all levels; share access failures captured for security monitoring
- **7-day backup retention** — Daily automatic backups at 01:00 UTC with tags copied to backups
- **Sunday maintenance** — Weekly maintenance window at Sunday 02:00 UTC

## What to Customize

- Replace placeholders: subnet, security group, KMS key ARN, AD domain details, credentials
- For production, switch to `domain_join_service_account_secret_arn` instead of inline `username`/`password`
- Increase `storage_capacity_gib` as data grows (can increase but never decrease)
- Increase `throughput_capacity` (next tiers: 512, 1024) if monitoring shows saturation
- Add `disk_iops_configuration` with `USER_PROVISIONED` mode for IOPS-heavy workloads
- Add `aliases` for user-friendly DNS mount points
- Increase `automatic_backup_retention_days` (max 90) for longer recovery windows
- Switch to `MULTI_AZ_1` when you need cross-AZ failover
