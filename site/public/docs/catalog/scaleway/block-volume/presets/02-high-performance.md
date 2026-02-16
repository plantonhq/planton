---
title: "High-Performance Block Volume"
description: "This preset creates a 100 GB Scaleway Block Storage volume with the high-performance 15,000 IOPS tier. This is the standard configuration for databases, search engines, and other I/O-intensive..."
type: "preset"
rank: "02"
presetSlug: "02-high-performance"
componentSlug: "block-volume"
componentTitle: "Block Volume"
provider: "scaleway"
icon: "package"
order: 2
---

# High-Performance Block Volume

This preset creates a 100 GB Scaleway Block Storage volume with the high-performance 15,000 IOPS tier. This is the standard configuration for databases, search engines, and other I/O-intensive workloads that require low-latency and high-throughput persistent storage.

## When to Use

- Database data directories (PostgreSQL, MySQL, MongoDB running on instances)
- Search engine indices (Elasticsearch, OpenSearch)
- Any workload where I/O performance is critical to application latency

## Key Configuration Choices

- **15k IOPS tier** (`performanceTier: sbs_15k`) -- 15,000 IOPS baseline; 3x the throughput of the standard tier, designed for database and analytics workloads
- **100 GB size** (`sizeGb: 100`) -- a common starting size for production databases; can be increased after creation (range: 5-10,240 GB)
- **Zonal** -- must be in the same zone as the instance it attaches to

## Placeholders to Replace

No placeholders -- this preset is ready to deploy as-is. Adjust `sizeGb` for your data volume requirements and `zone` for your target availability zone.

## Related Presets

- **01-standard** -- Use instead for general-purpose storage with moderate I/O needs at a lower cost
