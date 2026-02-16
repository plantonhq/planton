---
title: "Single-AZ Development FSx ONTAP"
description: "SINGLE_AZ_2 SSD file system with 1 TiB (1024 GiB) and 128 MB/s throughput. One HA pair. No automatic backups. The smallest and cheapest ONTAP configuration for development and testing."
type: "preset"
rank: "01"
presetSlug: "01-single-az-development"
componentSlug: "fsx-for-ontap"
componentTitle: "FSx for ONTAP"
provider: "aws"
icon: "package"
order: 1
---

# Single-AZ Development FSx ONTAP

SINGLE_AZ_2 SSD file system with 1 TiB (1024 GiB) and 128 MB/s throughput. One HA pair. No automatic backups. The smallest and cheapest ONTAP configuration for development and testing.

## When to Use

- Development and test environments for NFS/SMB/iSCSI workloads
- Proof-of-concept deployments to validate ONTAP access patterns
- CI/CD pipelines needing a shared file system temporarily
- Quick experiments where data loss is acceptable

## What It Configures

- **SINGLE_AZ_2** — Latest generation single-AZ deployment with in-place HA scale-out support
- **1024 GiB SSD** — Minimum storage capacity. Sub-millisecond latency
- **128 MB/s throughput** — Minimum throughput tier per HA pair
- **1 HA pair** — Single pair for cost efficiency
- **No backups** — `automatic_backup_retention_days: 0` disables daily FSx backups

## What to Customize

- Replace placeholders: `name`, `id`, `org`, `env`, and `subnet-0123456789abcdef0`
- Increase `storage_capacity_gib` (1024–1048576 GiB) for more space
- Increase `throughput_capacity_per_ha_pair` (next tiers: 256, 384, 512) for faster I/O
- Set `automatic_backup_retention_days: 7` if you need backup protection
- Add `security_group_ids` for network access control
- Add `ha_pairs: 2` or more for scale-out throughput (single-AZ only)
- Switch to `MULTI_AZ_2` when you need high availability
