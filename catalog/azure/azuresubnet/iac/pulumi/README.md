# AzureSubnet Pulumi Module

## Overview

This Pulumi module provisions an Azure subnet using the Azure Classic
provider (`pulumi-azure`). It creates a `network.Subnet` -- the workload
segment that partitions a virtual network's address space -- plus one
association resource per declared attachment (route table, network
security group, NAT gateway).

The subnet is an ARM child of its virtual network: the module derives the
resource group and network name from the parent network's ARM ID.
Attachments are declared subnet-side (matching Azure's model): one table,
group, or gateway serves many subnets, and each association is its own
ARM operation.

Everything except name and the parent network updates in place; those two
are the subnet's ARM identity, so changing either replaces the subnet and
everything deployed into it. A prefix in use cannot shrink, and toggling
`default_outbound_access_enabled` requires the subnet to be empty of VMs.
Subnets are not tracked ARM resources, so they carry no tags.

## Resources Created

- `network.Subnet` -- the subnet
- `network.SubnetRouteTableAssociation` / `SubnetNetworkSecurityGroupAssociation` / `SubnetNatGatewayAssociation` -- one per declared attachment

## Inputs

The module receives an `AzureSubnetIacInput` containing:

- `target.spec.virtual_network_id` -- ARM ID of the parent network (reference resolved to a literal by the platform); resource group and network name are parsed from it
- `target.spec.name` -- the subnet's name, unique within the network
- `target.spec.address_prefixes` XOR `target.spec.ip_address_pool` -- self-managed CIDR blocks, or Network Manager IPAM allocation (exactly one, enforced by spec validation)
- `target.spec.service_endpoints` / `service_endpoint_policy_ids` -- backbone routing to Azure services and its zero-trust narrowing
- `target.spec.delegations` -- hand the subnet to a PaaS service (e.g. `Microsoft.DBforPostgreSQL/flexibleServers`)
- `target.spec.private_endpoint_network_policies` / `private_link_service_network_policies_enabled` -- policy toggles; unset applies Azure's defaults
- `target.spec.default_outbound_access_enabled` -- Azure's historical default true; production subnets set false and route egress explicitly
- `target.spec.route_table_id` / `network_security_group_id` / `nat_gateway_id` -- the three attach seams, realized as association resources
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `subnet_id` | Full ARM ID of the subnet -- the join key downstream workloads deploy through |
| `subnet_name` | The subnet's name within its network |
| `address_prefixes` | The CIDR blocks actually assigned (IPAM allocations surface here) |
| `virtual_network_name` | Parent network name, derived from its ARM ID |
| `resource_group_name` | Parent network's resource group, derived from its ARM ID |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
