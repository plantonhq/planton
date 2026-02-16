---
title: "Preset: Multi-AZ High Availability"
description: "**Use case**: Mission-critical production workloads requiring automatic failover across availability zones, provisioned IOPS, storage quotas, and extended backup retention."
type: "preset"
rank: "03"
presetSlug: "03-multi-az-high-availability"
componentSlug: "fsx-for-openzfs"
componentTitle: "FSx for OpenZFS"
provider: "aws"
icon: "package"
order: 3
---

# Preset: Multi-AZ High Availability

**Use case**: Mission-critical production workloads requiring automatic failover across availability zones, provisioned IOPS, storage quotas, and extended backup retention.

## Configuration

- **Deployment type**: MULTI_AZ_1 — automatic failover with floating IP across 2 AZs
- **Storage**: 1024 GiB — production-scale for enterprise datasets
- **Throughput**: 1280 MB/s — high throughput for demanding workloads
- **Disk IOPS**: USER_PROVISIONED at 100,000 IOPS — guaranteed I/O performance
- **Compression**: ZSTD — maximizes effective capacity
- **Record size**: 128 KiB (default) — balanced for mixed workloads
- **NFS exports**: Restricted to VPC CIDR (10.0.0.0/16) with rw/crossmnt
- **Quotas**: Root user capped at 200 GiB, primary group at 500 GiB
- **Encryption**: Customer-managed KMS key
- **Backups**: 14-day retention with daily 03:00 UTC window
- **Maintenance**: Sunday at 02:00 UTC
- **Two subnets**: Active/standby across different AZs
- **Route tables**: Automatic route management for failover

## When to use

- Mission-critical databases and application data
- Financial services and regulated industries requiring HA
- Enterprise NFS workloads with strict availability SLAs
- Workloads that cannot tolerate any data center failure

## Cost considerations

MULTI_AZ_1 costs approximately 2x compared to SINGLE_AZ_2 due to data replication across AZs. USER_PROVISIONED IOPS at 100,000 adds additional cost beyond what AUTOMATIC mode provides. The 14-day backup retention covers two full business weeks of point-in-time recovery. Storage quotas help control costs by preventing uncontrolled growth from individual users/groups.
