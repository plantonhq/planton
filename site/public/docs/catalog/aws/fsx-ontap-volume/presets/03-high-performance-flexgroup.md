---
title: "High-Performance FlexGroup Volume"
description: "A FlexGroup volume distributed across multiple aggregates for maximum throughput. FlexGroup volumes stripe data across aggregates, enabling parallel I/O for large-scale workloads."
type: "preset"
rank: "03"
presetSlug: "03-high-performance-flexgroup"
componentSlug: "fsx-ontap-volume"
componentTitle: "FSx ONTAP Volume"
provider: "aws"
icon: "package"
order: 3
---

# High-Performance FlexGroup Volume

A FlexGroup volume distributed across multiple aggregates for maximum throughput. FlexGroup volumes stripe data across aggregates, enabling parallel I/O for large-scale workloads.

## When to use

- Data lakes with hundreds of TBs
- Genomics and bioinformatics pipelines
- Media rendering and post-production workflows
- Machine learning training data
- Any workload that benefits from parallel throughput

## Key settings

- **1 TB** initial size (thin-provisioned)
- **FLEXGROUP** volume style distributed across 2 aggregates
- **8 constituents per aggregate** (16 total member volumes for parallel I/O)
- **NONE** tiering — all data stays on primary SSD for consistent latency
- **Storage efficiency** still enabled (dedup across constituents)
