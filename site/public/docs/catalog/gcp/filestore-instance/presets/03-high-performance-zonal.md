---
title: "Preset: High Performance Zonal"
description: "**Tier**: ZONAL (modern SSD with IOPS tuning) **Use case**: Performance-sensitive workloads requiring high throughput"
type: "preset"
rank: "03"
presetSlug: "03-high-performance-zonal"
componentSlug: "filestore-instance"
componentTitle: "Filestore Instance"
provider: "gcp"
icon: "package"
order: 3
---

# Preset: High Performance Zonal

**Tier**: ZONAL (modern SSD with IOPS tuning)
**Use case**: Performance-sensitive workloads requiring high throughput

## What This Preset Provides

A high-performance Filestore instance optimized for throughput:

- **ZONAL tier**: modern SSD-backed storage with performance tuning support
- **2.5 TiB capacity**: SSD-backed with room for working datasets
- **20,000 fixed IOPS**: guaranteed IOPS regardless of capacity changes
- **CMEK encryption**: customer-managed encryption keys for compliance
- **PRIVATE_SERVICE_ACCESS**: secure network connectivity

## When to Use

- Media rendering and transcoding pipelines
- EDA (Electronic Design Automation) workloads
- Genomics and scientific computing
- Machine learning training data staging
- Any workload where NFS IOPS is the bottleneck

## When NOT to Use

- Archival or cold storage (use GCS or BASIC_HDD instead)
- Workloads that don't benefit from IOPS tuning (use BASIC_SSD)
- Multi-zone HA required (use ENTERPRISE or REGIONAL)
