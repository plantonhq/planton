---
title: "Presets"
description: "Ready-to-deploy configuration presets for FSx Lustre File System"
type: "preset-list"
componentSlug: "fsx-lustre-file-system"
componentTitle: "FSx Lustre File System"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-scratch-development"
    rank: "01"
    title: "Scratch Development FSx Lustre"
    excerpt: "SCRATCH_2 SSD file system with 1200 GiB — the smallest and cheapest Lustre configuration. No data replication, no backups. Temporary storage for fast processing."
  - slug: "02-persistent-high-throughput"
    rank: "02"
    title: "Persistent High Throughput FSx Lustre"
    excerpt: "PERSISTENT_2 SSD file system with 2400 GiB and 1000 MB/s/TiB throughput. Designed for production ML training, HPC simulations, and video rendering workloads that demand maximum I/O performance."
  - slug: "03-persistent-capacity-datalake"
    rank: "03"
    title: "Persistent Capacity Data Lake FSx Lustre"
    excerpt: "PERSISTENT_1 HDD file system with 6000 GiB and 12 MB/s/TiB throughput. Optimized for large-capacity, cost-effective storage where sequential throughput matters more than latency."
---

# FSx Lustre File System Presets

Ready-to-deploy configuration presets for FSx Lustre File System. Each preset is a complete manifest you can copy, customize, and deploy.
