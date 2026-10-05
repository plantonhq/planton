# AzureSubnet

## Overview

`AzureSubnet` provisions an Azure subnet: the workload segment that
partitions a virtual network's address space, and the composition hub of
Azure networking. Everything that deploys into a network -- AKS clusters,
databases, private endpoints, load balancers, VMs, function and web apps --
lands in a subnet by referencing its ID, and the network's building blocks
(route tables, network security groups, NAT gateways) attach themselves TO
the subnet.

## The Attach Model

Azure attaches routing, filtering, and egress to subnets -- not the other
way around. This kind models that faithfully with three optional
references:

- **`route_table_id`** -- steer the subnet's traffic through user-defined
  routes (firewall egress, forced tunneling)
- **`network_security_group_id`** -- filter the subnet's traffic with an
  NSG's rules
- **`nat_gateway_id`** -- give the subnet stable, scalable outbound
  connectivity through a NAT gateway's public addresses

Attachments are declared subnet-side because that is Azure's own shape: one
table, group, or gateway serves many subnets, while a subnet carries at
most one of each. Detaching is just removing the field.

## Key Features

- **Multi-prefix and dual-stack** -- `address_prefixes` is a first-class
  list; an IPv4 and an IPv6 block ride side by side
- **IPAM-delegated allocation** -- alternatively, delegate the CIDR to an
  Azure Network Manager IPAM pool (`ip_address_pool`); the provisioned
  range surfaces in the outputs
- **Service delegations** -- hand the subnet to a PaaS service (PostgreSQL
  Flexible Server, Container Apps, App Service VNet integration)
- **Service endpoints and endpoint policies** -- backbone routing to Azure
  services, optionally narrowed to specific resources
- **Private endpoint / private link policy control** -- the full ARM policy
  surface, including the granular NSG-only and route-table-only modes
- **Explicit egress posture** -- `default_outbound_access_enabled: false`
  turns off Azure's retiring implicit outbound access in favor of the NAT
  gateway or route-table egress you declare

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `virtual_network_id` | StringValueOrRef | Yes | Parent network's ARM ID (defaults to an AzureVirtualNetwork reference); RG and network name are derived from it |
| `name` | string | Yes | Subnet name, unique within the network (1-80 chars) |
| `address_prefixes` | list(string) | XOR | Self-managed CIDR blocks (exactly one of this or `ip_address_pool`) |
| `ip_address_pool` | object | XOR | Network Manager IPAM pool allocation (`id`, `number_of_ip_addresses`) |
| `service_endpoints` | list(string) | No | Azure services reachable over the backbone (e.g. "Microsoft.Storage") |
| `service_endpoint_policy_ids` | list(string) | No | Endpoint policies narrowing the endpoints' reach |
| `delegations` | list | No | PaaS service delegations (`name`, `service_name`, `actions`) |
| `private_endpoint_network_policies` | enum | No | ENABLED / NETWORK_SECURITY_GROUP_ENABLED / ROUTE_TABLE_ENABLED (unset = Azure's Disabled) |
| `private_link_service_network_policies_enabled` | bool | No | Azure default: true |
| `default_outbound_access_enabled` | bool | No | Azure default: true; set false for explicit egress |
| `sharing_scope` | enum | No | TENANT (requires default outbound access disabled) |
| `route_table_id` | StringValueOrRef | No | Route table to attach (defaults to an AzureRouteTable reference) |
| `network_security_group_id` | StringValueOrRef | No | NSG to attach (defaults to an AzureNetworkSecurityGroup reference) |
| `nat_gateway_id` | StringValueOrRef | No | NAT gateway to attach (defaults to an AzureNatGateway reference) |

Subnets have no `region` (they inherit the network's) and no `tags`
(subnets are not tracked ARM resources).

## Outputs

| Output | Description |
|--------|-------------|
| `subnet_id` | Full ARM ID -- the join key every network-attached resource consumes |
| `subnet_name` | The subnet's name within its network |
| `address_prefixes` | The ACTUAL assigned CIDRs (IPAM-provisioned when a pool allocates) |
| `virtual_network_name` | Parent network name, derived from the referenced ARM ID |
| `resource_group_name` | Resource group name, derived from the referenced ARM ID |

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureSubnet
metadata:
  name: app-subnet
  org: mycompany
  env: production
spec:
  virtualNetworkId:
    valueFrom:
      name: prod-vnet
  name: app
  addressPrefixes:
    - "10.0.1.0/24"
  serviceEndpoints:
    - Microsoft.Storage
  defaultOutboundAccessEnabled: false
  networkSecurityGroupId:
    valueFrom:
      name: app-tier-nsg
  natGatewayId:
    valueFrom:
      name: prod-egress
```

## Lifecycle Notes

- Address prefixes, endpoints, policies, delegations, and all three
  attachments update **in place**
- Name and the parent network are the subnet's ARM identity; changing
  either **replaces the subnet** and everything deployed into it
- A prefix in use by deployed resources cannot shrink; toggling
  `default_outbound_access_enabled` requires the subnet to be empty of VMs
- Azure reserves 5 IPs per subnet (the first four and the last)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
