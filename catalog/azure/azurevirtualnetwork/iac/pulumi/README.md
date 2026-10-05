# AzureVirtualNetwork Pulumi Module

## Overview

This Pulumi module provisions an Azure Virtual Network using the Azure
Classic provider (`pulumi-azure`). It creates a single
`network.VirtualNetwork` carrying the address space and network-wide
policy: multi-CIDR (or IPAM-delegated) addressing, custom DNS, BGP
community, DDoS Protection Plan attachment, VM-to-VM encryption, flow
timeout, and edge-zone placement.

Address space, DNS servers, BGP community, DDoS attachment, encryption,
flow timeout, and tags update in place; name, region, resource group, and
edge zone are the network's ARM identity, so changing any of them replaces
the network and everything inside it.

The network is deliberately just the network: subnets (`AzureSubnet`),
outbound NAT (`AzureNatGateway`), and private DNS attachments
(`AzurePrivateDnsZoneVirtualNetworkLink`) are separate composable
resources referencing this network's outputs.

## Resources Created

- `network.VirtualNetwork` -- the virtual network

## Inputs

The module receives an `AzureVirtualNetworkIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the network's ARM identity (references resolved to literals by the platform)
- `target.spec.address_spaces` OR `target.spec.ip_address_pools` -- exactly one address source (self-managed CIDRs or Network Manager IPAM delegation)
- `target.spec.dns_servers` -- custom resolvers (empty = Azure's default resolver)
- `target.spec.bgp_community` / `ddos_protection_plan` / `encryption` / `flow_timeout_in_minutes` / `private_endpoint_vnet_policies` / `edge_zone` -- network-wide policy
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `virtual_network_id` | Full ARM ID of the network -- the join key for subnets, peerings, and DNS links |
| `virtual_network_name` | The network's name as deployed |
| `guid` | ARM's stable network GUID |
| `address_spaces` | The ACTUAL ranges carried (IPAM-provisioned when pools delegate allocation) |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
