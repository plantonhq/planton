---
title: "Standard Floating IP Association"
description: "This preset binds an existing floating IP to a port. It is the \"join\" resource that connects a pre-allocated floating IP (from `OpenStackFloatingIp`) to a port (from `OpenStackNetworkPort`),..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "floating-ip-associate"
componentTitle: "Floating IP Associate"
provider: "openstack"
icon: "package"
order: 1
---

# Standard Floating IP Association

This preset binds an existing floating IP to a port. It is the "join" resource that connects a pre-allocated floating IP (from `OpenStackFloatingIp`) to a port (from `OpenStackNetworkPort`), providing external connectivity to whatever is attached to that port.

## When to Use

- InfraCharts where floating IP allocation and association are separate DAG nodes
- Associating a pre-existing or pre-reserved floating IP with a newly created port
- Re-pointing a floating IP from one port to another (only `portId` can be updated; changing `floatingIp` recreates)

## Key Configuration Choices

- **Join resource** -- binds a floating IP address to a port; no additional configuration
- **References the IP address** (`floatingIp`) -- this is the actual IP address string (e.g., `203.0.113.42`), not the floating IP UUID

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<floating-ip-address>` | The floating IP address (e.g., `203.0.113.42`) | OpenStack console or `OpenStackFloatingIp` status outputs (`address` field) |
| `<port-id>` | ID of the port to associate the floating IP with | OpenStack console or `OpenStackNetworkPort` status outputs |
