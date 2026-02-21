---
title: "Spread Placement Group"
description: "This preset creates a placement group with the `spread` strategy, which guarantees that servers assigned to it run on different physical hosts within a Hetzner Cloud datacenter. If the hypervisor..."
type: "preset"
rank: "01"
presetSlug: "01-spread"
componentSlug: "hetzner-cloud-placement-group"
componentTitle: "Hetzner Cloud Placement Group"
provider: "hetznercloud"
icon: "package"
order: 1
---

# Spread Placement Group

This preset creates a placement group with the `spread` strategy, which guarantees that servers assigned to it run on different physical hosts within a Hetzner Cloud datacenter. If the hypervisor hosting one server fails, no other server in the group is affected.

Hetzner Cloud currently offers only the `spread` strategy -- there is no `cluster` or `partition` equivalent. A single placement group supports up to 10 servers.

## When to Use

- High-availability database replicas (e.g., PostgreSQL primary + standby on separate hosts)
- Distributed application server fleets where a single host failure must not take down multiple instances
- Any workload where physical host-level fault isolation is a deployment requirement

## Key Configuration Choices

- **`type: spread`** -- the only placement strategy Hetzner Cloud supports; set explicitly so the manifest is self-documenting

## Placeholders to Replace

No placeholders -- this preset is ready to deploy after setting `metadata.name` to the desired placement group name.
