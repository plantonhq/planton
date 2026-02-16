---
title: "Standard Block Volume"
description: "This preset creates a 20 GB Scaleway Block Storage volume with the standard 5,000 IOPS performance tier. Block volumes are network-attached persistent storage that can be attached to instances and..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "block-volume"
componentTitle: "Block Volume"
provider: "scaleway"
icon: "package"
order: 1
---

# Standard Block Volume

This preset creates a 20 GB Scaleway Block Storage volume with the standard 5,000 IOPS performance tier. Block volumes are network-attached persistent storage that can be attached to instances and survive instance termination. This is the most common volume configuration for general-purpose persistent storage.

## When to Use

- Persistent data storage for application servers (logs, uploads, databases)
- Storage that needs to survive instance replacement or termination
- General-purpose workloads with moderate I/O requirements

## Key Configuration Choices

- **5k IOPS tier** (`performanceTier: sbs_5k`) -- 5,000 IOPS baseline; sufficient for most application workloads including light database use
- **20 GB size** (`sizeGb: 20`) -- starting size; can be increased after creation (range: 5-10,240 GB)
- **Zonal** -- block volumes are zonal resources; must be in the same zone as the instance they attach to

## Placeholders to Replace

No placeholders -- this preset is ready to deploy as-is. Adjust `sizeGb` and `zone` to match your requirements.

## Related Presets

- **02-high-performance** -- Use instead for database volumes or I/O-intensive workloads requiring 15,000 IOPS
