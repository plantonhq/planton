---
title: "Standard Router Interface"
description: "This preset attaches a subnet to a router, enabling Layer 3 routing for that subnet. Without a router interface, a subnet is an isolated Layer 2 domain with no connectivity to other subnets or..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "router-interface"
componentTitle: "Router Interface"
provider: "openstack"
icon: "package"
order: 1
---

# Standard Router Interface

This preset attaches a subnet to a router, enabling Layer 3 routing for that subnet. Without a router interface, a subnet is an isolated Layer 2 domain with no connectivity to other subnets or external networks.

## When to Use

- Every subnet that needs routing (which is most subnets)
- Connecting a workload subnet to an edge router for internet access
- Connecting multiple subnets to an internal router for east-west traffic

## Key Configuration Choices

- **Join resource** -- this is a simple binding between a router and a subnet with no additional configuration
- **ForceNew** -- all fields are immutable; changing either reference recreates the interface

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<router-id>` | ID of the router to attach the subnet to | OpenStack console or `OpenStackRouter` status outputs |
| `<subnet-id>` | ID of the subnet to connect to the router | OpenStack console or `OpenStackSubnet` status outputs |
