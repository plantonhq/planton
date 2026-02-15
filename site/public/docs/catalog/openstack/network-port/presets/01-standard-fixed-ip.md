---
title: "Standard Port with Fixed IP"
description: "This preset creates a port on a network with an IP auto-assigned from a specific subnet. The port gets the project's default security group applied automatically. This is the most common port..."
type: "preset"
rank: "01"
presetSlug: "01-standard-fixed-ip"
componentSlug: "network-port"
componentTitle: "Network Port"
provider: "openstack"
icon: "package"
order: 1
---

# Standard Port with Fixed IP

This preset creates a port on a network with an IP auto-assigned from a specific subnet. The port gets the project's default security group applied automatically. This is the most common port configuration -- suitable for attaching to instances, load balancers, or other network-consuming resources.

## When to Use

- Pre-provisioning a network identity before creating an instance
- Attaching an instance to a specific subnet with a stable IP address
- Creating a port for floating IP association or load balancer VIP

## Key Configuration Choices

- **Explicit subnet** (`fixedIps[].subnetId`) -- ensures the IP comes from the specified subnet rather than OpenStack auto-selecting
- **Auto-assigned IP** -- OpenStack picks an available IP from the subnet's allocation pool
- **Default security group** -- project's default SG is applied automatically (omitting `securityGroupIds` triggers this behavior)
- **Admin state up** -- default (true), port forwards traffic immediately

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<network-id>` | ID of the network to create the port on | OpenStack console or `OpenStackNetwork` status outputs |
| `<subnet-id>` | ID of the subnet to allocate an IP from | OpenStack console or `OpenStackSubnet` status outputs |
