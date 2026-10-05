# AzureNetworkInterface Pulumi Module

## Overview

This Pulumi module provisions an Azure Network Interface using the Azure
Classic provider (`pulumi-azure`). It creates a `network.NetworkInterface`
-- the attachment point that gives a virtual machine its presence in a
subnet -- plus an association resource for the NIC-level network security
group and one per application security group membership.

The NIC is the attachment point, not the machine. An `AzureVirtualMachine`
consumes this module's `network_interface_id` output via its
`network_interface_ids` (a VM references its NICs, never contains them),
each IP configuration deploys into a referenced `AzureSubnet` and may
front a referenced `AzurePublicIp`, and the NIC-level NSG is the
per-workload complement to subnet-level filtering (inbound traffic must
pass the subnet NSG then the NIC NSG when both are attached).

Name, region, and edge zone are the NIC's identity -- changing any of
them replaces the NIC, detaching it from its VM. Everything else
(configurations, DNS, acceleration, forwarding, associations, tags)
updates in place. The MAC address is assigned when the NIC attaches to a
running VM, not at creation.

## Resources Created

- `network.NetworkInterface` -- the NIC itself
- `network.NetworkInterfaceSecurityGroupAssociation` -- when a NIC-level NSG is referenced
- `network.NetworkInterfaceApplicationSecurityGroupAssociation` -- one per application security group membership

## Inputs

The module receives an `AzureNetworkInterfaceIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the NIC's ARM identity (references resolved to literals by the platform)
- `target.spec.ip_configurations` -- at least one; each places a private address in a subnet (DYNAMIC or STATIC allocation, IPv4 or IPv6) and may front a public IP. With multiple, the first must be primary (ARM's contract, spec-enforced)
- `target.spec.dns_servers` / `target.spec.internal_dns_name_label` -- per-NIC DNS override and the VNet-internal resolution label
- `target.spec.accelerated_networking_enabled` -- SR-IOV; Azure defaults to false, enable on every supported VM size
- `target.spec.ip_forwarding_enabled` -- forwarding of traffic not addressed to the NIC; network virtual appliances only
- `target.spec.auxiliary_mode` / `target.spec.auxiliary_sku` -- preview NVA acceleration (subscription must be enrolled); both or neither, unspecified sends nothing
- `target.spec.edge_zone` -- Edge Zone pinning; fixed at creation
- `target.spec.network_security_group_id` -- the NIC-level NSG, resolved to an ARM ID; realized as an association resource
- `target.spec.application_security_group_ids` -- ASG memberships, as ARM IDs; one association resource each
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `network_interface_id` | Full ARM ID of the NIC -- the join key `AzureVirtualMachine` attaches through |
| `network_interface_name` | The NIC's name as deployed |
| `private_ip_address` | The primary configuration's private IP address |
| `private_ip_addresses` | All configurations' private addresses, in configuration order |
| `mac_address` | The NIC's MAC address (populated once attached to a running VM) |
| `internal_domain_name_suffix` | The DNS suffix completing `internal_dns_name_label` into a resolvable FQDN |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
