# AzureNatGateway Pulumi Module

## Overview

This Pulumi module provisions an Azure NAT Gateway using the Azure Classic
provider (`pulumi-azure`). It creates a `network.NatGateway` -- the managed
SNAT service that gives every workload in its attached subnets stable,
scalable outbound connectivity -- plus one association resource per
referenced public IP and public IP prefix.

The gateway is just the gateway. The addresses it SNATs through are
referenced first-class `AzurePublicIp` / `AzurePublicIpPrefix` resources
(each adds 64,512 SNAT ports; a /28 prefix scales that by 16 in one
allowlistable range), and the subnets it serves attach themselves via
`AzureSubnet`'s `nat_gateway_id` -- matching Azure's model, so one gateway
serves many subnets without listing them. A gateway with no associated
addresses deploys but cannot translate anything.

Idle timeout, tags, and the IP/prefix associations update in place. Name,
SKU, and zone are the gateway's identity -- changing any of them replaces
it, briefly interrupting egress for every attached subnet.

## Resources Created

- `network.NatGateway` -- the gateway itself
- `network.NatGatewayPublicIpAssociation` -- one per referenced public IP
- `network.NatGatewayPublicIpPrefixAssociation` -- one per referenced prefix

## Inputs

The module receives an `AzureNatGatewayIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the gateway's ARM identity (references resolved to literals by the platform)
- `target.spec.sku_name` -- STANDARD (Azure's default, zonal) or STANDARD_V2 (zone-redundant automatically, requires StandardV2 addresses and empty zones)
- `target.spec.idle_timeout_in_minutes` -- SNAT port hold time for idle TCP connections, 4-120; Azure defaults to 4
- `target.spec.zones` -- optional single zone to pin a STANDARD gateway to; must be empty for STANDARD_V2
- `target.spec.public_ip_ids` / `target.spec.public_ip_prefix_ids` -- ARM IDs of the addresses and ranges to SNAT through
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `nat_gateway_id` | Full ARM ID of the gateway -- the join key subnets attach through |
| `nat_gateway_name` | The gateway's name as deployed |
| `resource_guid` | The immutable GUID ARM assigns the gateway (billing/support correlation) |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
