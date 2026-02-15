---
title: "Standard DHCP Subnet"
description: "This preset creates an IPv4 subnet with DHCP enabled and Google public DNS servers. OpenStack automatically assigns the first usable IP as the gateway and allocates the remaining range via DHCP. This..."
type: "preset"
rank: "01"
presetSlug: "01-standard-dhcp"
componentSlug: "subnet"
componentTitle: "Subnet"
provider: "openstack"
icon: "package"
order: 1
---

# Standard DHCP Subnet

This preset creates an IPv4 subnet with DHCP enabled and Google public DNS servers. OpenStack automatically assigns the first usable IP as the gateway and allocates the remaining range via DHCP. This is the most common subnet configuration for workload connectivity.

## When to Use

- Standard workload subnets that need automatic IP assignment
- Application networks where instances get IPs via DHCP
- Any subnet that will be attached to a router for external connectivity

## Key Configuration Choices

- **IPv4** (`ipVersion: 4`) -- default, suitable for most deployments
- **DHCP enabled** (`enableDhcp: true`) -- default, instances receive IPs automatically
- **Auto gateway** -- OpenStack assigns the first usable IP (e.g., 192.168.1.1) as the gateway
- **Google DNS** (`8.8.8.8`, `8.8.4.4`) -- reliable public DNS pushed to instances via DHCP
- **/24 CIDR** -- 254 usable addresses, standard for workload subnets

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<network-id>` | ID of the network this subnet belongs to | OpenStack console or `OpenStackNetwork` status outputs |
