# AzurePrivateDnsZone Pulumi Module

## Overview

This Pulumi module provisions an Azure Private DNS zone using the Azure
Classic provider (`pulumi-azure`). It creates a single `privatedns.Zone`
-- a global record container -- with optional SOA customization and
governance tags.

Tags update in place; the zone's name is its ARM identity, so renaming
replaces the zone and every record in it. The SOA record is written at
creation and cannot be customized afterwards.

The zone is deliberately just the zone: which networks can resolve it is
declared through `AzurePrivateDnsZoneVirtualNetworkLink` resources
referencing this zone's `zone_id` output -- one link per network. A zone
with no links answers nobody.

## Resources Created

- `privatedns.Zone` -- the private DNS zone

## Inputs

The module receives an `AzurePrivateDnsZoneIacInput` containing:

- `target.spec.resource_group` -- the zone's resource group (references resolved to a literal by the platform)
- `target.spec.name` -- the zone's DNS domain name (privatelink zone name or custom domain)
- `target.spec.soa_record` -- optional SOA customization (email + timers; unset timers fall back to Azure's defaults)
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `zone_id` | Full ARM ID of the zone -- the join key for links, private endpoints, and databases |
| `zone_name` | The zone's DNS name as deployed |
| `resource_group_name` | The zone's resource group |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
