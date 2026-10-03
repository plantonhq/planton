# AzurePrivateDnsZoneVirtualNetworkLink Pulumi Module

## Overview

This Pulumi module provisions a private DNS zone virtual network link
using the Azure Classic provider (`pulumi-azure`). It creates a single
`privatedns.ZoneVirtualNetworkLink`: the attachment that makes the
referenced zone resolvable from the referenced network.

`registration_enabled`, `resolution_policy`, and tags update in place;
name, zone, and network are the link's ARM identity, so changing any of
them replaces the link (a brief resolution gap for the affected network,
nothing else). Azure allows only ONE registration-enabled link per
network.

The module derives the zone's name and resource group from the referenced
zone's ARM ID -- the SDK takes them as separate arguments even though the
parent ID already carries them, and deriving them means they can never
disagree with the referenced zone.

## Resources Created

- `privatedns.ZoneVirtualNetworkLink` -- the zone-network attachment

## Inputs

The module receives an `AzurePrivateDnsZoneVirtualNetworkLinkIacInput` containing:

- `target.spec.name` -- the link's name under the parent zone
- `target.spec.private_dns_zone_id` -- the parent zone's ARM ID (references resolved to a literal by the platform; zone name + resource group derived from it)
- `target.spec.virtual_network_id` -- the network gaining resolution (references resolved to a literal)
- `target.spec.registration_enabled` -- VM A-record auto-registration (default false)
- `target.spec.resolution_policy` -- optional; unset lets Azure apply its per-zone-type default
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `link_id` | Full ARM ID of the link |
| `link_name` | The link's name as deployed |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
