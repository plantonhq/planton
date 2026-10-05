# AzureNetworkSecurityGroup Pulumi Module

## Overview

This Pulumi module provisions an Azure Network Security Group using the
Azure Classic provider (`pulumi-azure`). It creates a single
`network.NetworkSecurityGroup` with its security rules managed INLINE on
the group -- the inline form carries the full application-security-group
ID lists that the pinned SDK's standalone rule type flattens. (The
Terraform module manages the same rules as standalone resources; both
engines put the identical rule set into ARM.)

Rules and tags update in place and take effect immediately for every
subnet and NIC the group guards. Name, region, and resource group are the
group's ARM identity; changing any of them replaces it, detaching it from
every subnet until re-attached.

A group with no rules is meaningful: Azure's implicit defaults then govern
(allow VNet-internal and load-balancer traffic, deny other inbound, allow
all outbound). The subnet-side attachment is deliberately not modeled
here: a subnet declares which NSG guards it (AzureSubnet's
`network_security_group_id`), so one group serves many subnets.

## Resources Created

- `network.NetworkSecurityGroup` -- the group with inline rules

## Inputs

The module receives an `AzureNetworkSecurityGroupIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the group's ARM identity (references resolved to literals by the platform)
- `target.spec.security_rules` -- 5-tuple filters with `priority` (100-4096, lowest evaluates first), `direction` (INBOUND/OUTBOUND), `access` (ALLOW/DENY), and `protocol` (ANY/TCP/UDP/ICMP/AH/ESP, where ANY is ARM's "*"); ports and each address side take exactly one form (single prefix, prefix list, or application security group IDs), with unset meaning any
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `network_security_group_id` | Full ARM ID of the group -- the join key subnets attach through |
| `network_security_group_name` | The group's name as deployed |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
