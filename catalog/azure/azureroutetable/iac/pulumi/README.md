# AzureRouteTable Pulumi Module

## Overview

This Pulumi module provisions an Azure route table using the Azure Classic
provider (`pulumi-azure`). It creates a single `network.RouteTable` with
its user-defined routes managed inline (a route has no life of its own in
Azure) and BGP propagation control.

Routes, BGP propagation, and tags update in place -- and take effect
immediately for every subnet attached to the table. Name, region, and
resource group are the table's ARM identity; changing any of them replaces
the table, detaching it from every subnet until re-attached.

The subnet-side attachment is deliberately not modeled here: a subnet
declares which route table it uses (matching Azure's model), so one table
serves many subnets without listing them.

## Resources Created

- `network.RouteTable` -- the route table with inline routes

## Inputs

The module receives an `AzureRouteTableIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the table's ARM identity (references resolved to literals by the platform)
- `target.spec.routes` -- the user-defined routes; VIRTUAL_APPLIANCE routes carry the appliance IP (pairing enforced by spec validation)
- `target.spec.bgp_route_propagation_enabled` -- Azure defaults to true; false is the forced-tunneling hardening
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `route_table_id` | Full ARM ID of the table -- the join key subnets attach through |
| `route_table_name` | The table's name as deployed |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
