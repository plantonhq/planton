---
title: "Single-AZ Production FSx ONTAP"
description: "SINGLE_AZ_2 SSD file system with 2 TiB (2048 GiB) and 512 MB/s throughput. One HA pair. Customer-managed KMS encryption. 7-day automatic backups at 05:00 UTC. Production-ready configuration for..."
type: "preset"
rank: "02"
presetSlug: "02-single-az-production"
componentSlug: "fsx-for-ontap"
componentTitle: "FSx for ONTAP"
provider: "aws"
icon: "package"
order: 2
---

# Single-AZ Production FSx ONTAP

SINGLE_AZ_2 SSD file system with 2 TiB (2048 GiB) and 512 MB/s throughput. One HA pair. Customer-managed KMS encryption. 7-day automatic backups at 05:00 UTC. Production-ready configuration for non-HA workloads.

## When to Use

- Production workloads that can tolerate single-AZ availability
- Database storage (Oracle, SAP, SQL Server) on shared NFS/iSCSI
- VMware Cloud on AWS datastores
- Enterprise file shares with compliance requirements (encryption, backups)

## What It Configures

- **SINGLE_AZ_2** — Latest generation single-AZ deployment
- **2048 GiB SSD** — 2 TiB storage. Sub-millisecond latency
- **512 MB/s throughput** — Production-grade throughput tier
- **1 HA pair** — Standard redundancy within the AZ
- **Customer-managed KMS** — Encryption at rest with your key
- **7-day backups** — Daily automatic backups at 05:00 UTC
- **Copy tags to backups** — Cost allocation and resource tracking
- **Weekly maintenance** — Sunday at 02:00 UTC

## What to Customize

- Replace placeholders: `name`, `id`, `org`, `env`, `subnet-0123456789abcdef0`, `sg-0123456789abcdef0`, and KMS key ARN
- Increase `storage_capacity_gib` for larger datasets
- Increase `throughput_capacity_per_ha_pair` (768, 1024, 1536, 2048) for higher I/O
- Add `ha_pairs: 2` or more for scale-out throughput
- Adjust `automatic_backup_retention_days` (up to 90) for longer retention
- Use `valueFrom` references to wire AwsVpc, AwsSecurityGroup, and AwsKmsKey
- Switch to preset 03 for multi-AZ high availability
