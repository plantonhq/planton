---
title: "General Purpose ONTAP Volume"
description: "A standard read-write ONTAP volume suitable for most workloads. Mounted at `/data` with UNIX security style, storage efficiency enabled (deduplication, compression, compaction), and AUTO tiering with..."
type: "preset"
rank: "01"
presetSlug: "01-general-purpose"
componentSlug: "fsx-ontap-volume"
componentTitle: "FSx ONTAP Volume"
provider: "aws"
icon: "package"
order: 1
---

# General Purpose ONTAP Volume

A standard read-write ONTAP volume suitable for most workloads. Mounted at `/data` with UNIX security style, storage efficiency enabled (deduplication, compression, compaction), and AUTO tiering with a 31-day cooling period to optimize costs.

## When to use

- NFS-based application data volumes
- Shared file storage for Linux workloads
- General-purpose file shares where cost optimization is desired

## Key settings

- **100 GB** initial size (thin-provisioned, grows as needed)
- **UNIX** security style for Linux/NFS clients
- **AUTO** tiering moves cold data to capacity pool after 31 days
- **Storage efficiency** reduces physical storage consumption
