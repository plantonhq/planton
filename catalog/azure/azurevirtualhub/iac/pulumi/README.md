# AzureVirtualHub Pulumi Module

## Overview

This Pulumi module provisions a Virtual Hub using the Azure Classic
provider (`pulumi-azure`): a `network.VirtualHub` plus its composed
routing children -- `network.VirtualHubRouteTable` (routes inline),
`network.RouteMap`, `network.BgpConnection`, and
`network.RoutingIntent`, each parented under the hub.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- The route map's SDK constructor is `NewRouteMapResource` -- the
  resource lives at a legacy "RouteMapResource" Go type name but
  creates the SAME ARM object as `azurerm_route_map`.
- The spec's optional fields apply ARM's defaults when unset (Standard
  tier, ExpressRoute preference, router capacity 2) through
  nil-handling helpers in `locals.go`, mirroring the Terraform module's
  null handling.
- Enum names map to ARM wire values through explicit switch helpers --
  the same name-to-wire tables the Terraform module carries in its
  locals.
- Composed children export name-keyed ID maps (`route_table_ids`,
  `route_map_ids`, `bgp_connection_ids`) so downstream components
  reference children by the spec's own names.

## Inputs

The module receives an `AzureVirtualHubIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the hub's ARM identity (references resolved to literals by the platform)
- `target.spec.virtual_wan_id` -- the WAN the hub belongs to
- `target.spec.address_prefix` -- the hub's private CIDR
- `target.spec.sku` / `hub_routing_preference` / `virtual_router_auto_scale_min_capacity` -- tier and router tuning
- `target.spec.routes` / `route_tables` / `route_maps` / `bgp_connections` / `routing_intent` -- the routing children
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `virtual_hub_id` | Full ARM ID -- what connections and gateways reference |
| `virtual_hub_name` | The hub's name |
| `default_route_table_id` | The built-in default route table's ARM ID |
| `virtual_router_asn` | The hub router's BGP ASN (always 65515) |
| `virtual_router_ips` | The hub router's peering IPv4 addresses |
| `route_table_ids` | Custom route table ARM IDs, keyed by name |
| `route_map_ids` | Route map ARM IDs, keyed by name |
| `bgp_connection_ids` | BGP connection ARM IDs, keyed by name |
| `routing_intent_id` | The routing intent's ARM ID ("" when none) |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
