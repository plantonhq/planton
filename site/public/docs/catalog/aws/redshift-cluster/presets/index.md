---
title: "Presets"
description: "Ready-to-deploy configuration presets for Redshift Cluster"
type: "preset-list"
componentSlug: "redshift-cluster"
componentTitle: "Redshift Cluster"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-single-node-dev"
    rank: "01"
    title: "Single-Node Development Cluster"
    excerpt: "This preset creates a single-node Redshift cluster on the dc2.large instance type for development and testing. The single-node topology combines the leader and compute roles on one node, keeping..."
  - slug: "02-multi-node-production"
    rank: "02"
    title: "Multi-Node Production Data Warehouse"
    excerpt: "This preset creates a 2-node RA3 Redshift cluster configured for production workloads. RA3 nodes decouple compute and storage by automatically tiering data between local SSD and Amazon S3, so you can..."
  - slug: "03-analytics-workload"
    rank: "03"
    title: "High-Performance Analytics Cluster"
    excerpt: "This preset creates a 4-node RA3 Redshift cluster sized for large-scale analytical workloads. The ra3.4xlarge nodes each provide 12 vCPUs, 96 GiB RAM, and managed storage that automatically tiers..."
---

# Redshift Cluster Presets

Ready-to-deploy configuration presets for Redshift Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
