# AzureLocalNetworkGateway Pulumi Module

## Overview

This Pulumi module provisions a local network gateway using the Azure
Classic provider (`pulumi-azure`). It creates a
`network.LocalNetworkGateway` -- Azure's description of the on-premises
side of a site-to-site VPN: the device's public endpoint and the
address space behind it. Nothing deploys on-premises; the object
provisions in seconds and costs nothing to keep.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- The endpoint contract (exactly one of address or FQDN) and the
  routing-source contract (static prefixes, BGP, or both) are
  spec-validated -- the module maps fields without re-checking.
- Empty optional fields are omitted rather than sent as zero values, so
  both engines produce identical ARM payloads.

## Inputs

The module receives an `AzureLocalNetworkGatewayIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the object's ARM identity (references resolved to literals by the platform)
- `target.spec.gateway_address` / `target.spec.gateway_fqdn` -- the device's public endpoint (exactly one)
- `target.spec.address_spaces` -- CIDRs Azure routes into the tunnel
- `target.spec.bgp_settings` -- the site's BGP speaker (ASN, tunnel-interior peering address, weight)
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `local_network_gateway_id` | Full ARM ID -- the join key connections reference |
| `local_network_gateway_name` | The description's name as deployed |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
