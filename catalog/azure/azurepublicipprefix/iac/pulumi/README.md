# AzurePublicIpPrefix Pulumi Module

## Overview

This Pulumi module provisions an Azure Public IP Prefix using the Azure
Classic provider (`pulumi-azure`). It creates a single
`network.PublicIpPrefix` -- a reserved, contiguous range of public IP
addresses that public IPs allocate from (`AzurePublicIp.public_ip_prefix_id`)
and NAT gateways associate for outbound SNAT
(`AzureNatGateway.public_ip_prefix_ids`).

The prefix is essentially immutable: everything except tags is fixed at
creation, and replacing it changes the ACTUAL reserved range -- everything
partners have allowlisted. Treat replacement as a coordinated migration,
never a casual update. The prefix cannot be deleted while any of its
addresses are in use.

Only explicit choices are sent to Azure: an unspecified prefix length,
version, SKU, or tier falls through to Azure's defaults, so an
unspecified spec deploys identically on both engines.

## Resources Created

- `network.PublicIpPrefix` -- the reserved address range

## Inputs

The module receives an `AzurePublicIpPrefixIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the prefix's ARM identity (references resolved to literals by the platform)
- `target.spec.prefix_length` -- CIDR length of the reserved range; Azure defaults to 28 (16 addresses), /31 reserves 2, /29 reserves 8; smaller numbers reserve bigger ranges and bill for every reserved address
- `target.spec.ip_version` -- IPV4 (Azure's default) or IPV6
- `target.spec.sku` / `target.spec.sku_tier` -- STANDARD/REGIONAL are Azure's defaults; STANDARD_V2 pairs with StandardV2 NAT gateways; GLOBAL is for cross-region load balancer frontends and requires STANDARD
- `target.spec.zones` -- availability zones anchoring the range; multiple zones make it zone-redundant
- `target.spec.custom_ip_prefix_id` -- optional BYOIP custom prefix to carve the range from
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins); the only field that updates in place
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `public_ip_prefix_id` | Full ARM ID of the prefix -- the join key public IPs and NAT gateways reference |
| `ip_prefix` | The actual reserved CIDR (e.g. `20.42.0.16/28`), known only after creation -- the value partners allowlist |
| `public_ip_prefix_name` | The prefix's name as deployed |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
