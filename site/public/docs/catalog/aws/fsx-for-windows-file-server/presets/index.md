---
title: "Presets"
description: "Ready-to-deploy configuration presets for FSx for Windows File Server"
type: "preset-list"
componentSlug: "fsx-for-windows-file-server"
componentTitle: "FSx for Windows File Server"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-single-az-development"
    rank: "01"
    title: "Single-AZ Development FSx Windows"
    excerpt: "SINGLE_AZ_2 SSD file system with 32 GiB and 32 MB/s throughput joined to an AWS Managed Microsoft AD. The smallest and cheapest Windows file system configuration. Backups disabled for cost savings."
  - slug: "02-single-az-production"
    rank: "02"
    title: "Single-AZ Production FSx Windows"
    excerpt: "SINGLE_AZ_2 SSD file system with 500 GiB and 256 MB/s throughput. Configured for production workloads with self-managed Active Directory, audit logging for compliance, customer-managed KMS..."
  - slug: "03-multi-az-high-availability"
    rank: "03"
    title: "Multi-AZ High Availability FSx Windows"
    excerpt: "MULTI_AZ_1 SSD file system with 1000 GiB, 512 MB/s throughput, and automatic cross-AZ failover. Mission-critical configuration with DNS aliases, provisioned IOPS (100,000), full audit logging on both..."
---

# FSx for Windows File Server Presets

Ready-to-deploy configuration presets for FSx for Windows File Server. Each preset is a complete manifest you can copy, customize, and deploy.
