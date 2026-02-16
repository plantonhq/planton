---
title: "Presets"
description: "Ready-to-deploy configuration presets for FSx ONTAP Volume"
type: "preset-list"
componentSlug: "fsx-ontap-volume"
componentTitle: "FSx ONTAP Volume"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-general-purpose"
    rank: "01"
    title: "General Purpose ONTAP Volume"
    excerpt: "A standard read-write ONTAP volume suitable for most workloads. Mounted at `/data` with UNIX security style, storage efficiency enabled (deduplication, compression, compaction), and AUTO tiering with..."
  - slug: "02-compliance-snaplock"
    rank: "02"
    title: "Compliance SnapLock Volume"
    excerpt: "An ONTAP volume with SnapLock COMPLIANCE for immutable record retention. Files committed to this volume become Write Once Read Many (WORM) — they cannot be modified or deleted by anyone until their..."
  - slug: "03-high-performance-flexgroup"
    rank: "03"
    title: "High-Performance FlexGroup Volume"
    excerpt: "A FlexGroup volume distributed across multiple aggregates for maximum throughput. FlexGroup volumes stripe data across aggregates, enabling parallel I/O for large-scale workloads."
---

# FSx ONTAP Volume Presets

Ready-to-deploy configuration presets for FSx ONTAP Volume. Each preset is a complete manifest you can copy, customize, and deploy.
