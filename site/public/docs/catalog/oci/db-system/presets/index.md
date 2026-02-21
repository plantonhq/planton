---
title: "Presets"
description: "Ready-to-deploy configuration presets for DB System"
type: "preset-list"
componentSlug: "db-system"
componentTitle: "DB System"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-single-node-vm"
    rank: "01"
    title: "Single-Node VM"
    excerpt: "This preset creates a single-node Oracle Database System on a VM.Standard.E4.Flex shape with Standard Edition, automatic backups, and a pluggable database. This is the baseline for running a managed..."
  - slug: "02-two-node-rac"
    rank: "02"
    title: "Two-Node RAC"
    excerpt: "This preset creates a two-node Real Application Clusters (RAC) Oracle Database System for high availability. The nodes are distributed across fault domains for infrastructure-level resilience, with..."
---

# DB System Presets

Ready-to-deploy configuration presets for DB System. Each preset is a complete manifest you can copy, customize, and deploy.
