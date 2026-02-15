---
title: "Anti-Affinity Server Group"
description: "This preset creates a server group with the anti-affinity policy. Instances placed in this group are scheduled on different physical hypervisors, maximizing fault tolerance. If a hypervisor fails,..."
type: "preset"
rank: "01"
presetSlug: "01-anti-affinity"
componentSlug: "server-group"
componentTitle: "Server Group"
provider: "openstack"
icon: "package"
order: 1
---

# Anti-Affinity Server Group

This preset creates a server group with the anti-affinity policy. Instances placed in this group are scheduled on different physical hypervisors, maximizing fault tolerance. If a hypervisor fails, only one member of the group is affected.

## When to Use

- HA database clusters (primary and replica on separate hosts)
- Load-balanced application tiers where host-level failure should not take down all instances
- Any workload where hardware fault isolation is required

## Key Configuration Choices

- **Anti-affinity** (`policy: anti-affinity`) -- strict scheduling constraint; instances will not co-locate on the same hypervisor
- **Immutable** -- all fields are ForceNew; changing the policy requires recreating the group and re-assigning instances

## Placeholders to Replace

No placeholders -- this preset is deployable as-is after setting `metadata.name`.

## Related Presets

- **02-affinity** -- Use instead when instances should be co-located on the same hypervisor (low-latency communication)
