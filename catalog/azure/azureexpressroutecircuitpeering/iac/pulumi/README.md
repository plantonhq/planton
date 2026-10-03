# AzureExpressRouteCircuitPeering Pulumi Module

## Overview

This Pulumi module provisions an ExpressRoute circuit peering using the
Azure Classic provider (`pulumi-azure`). It creates a
`network.ExpressRouteCircuitPeering` plus one
`network.ExpressRouteCircuitConnection` child per spec `connections`
entry, parented to the peering -- Global Reach links to other circuits'
private peerings.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- The type-dependent contracts (route filter / Microsoft config /
  IPv6 / connections per peering type) are spec-validated -- the module
  maps fields without re-checking.
- The near side of every Global Reach connection is wired to the
  peering this module just created (`createdPeering.ID()`); only the
  far side is user surface -- identical to the Terraform module's
  `peering_id` wiring.
- The spec's peering-type enum maps to ARM wire values (which are also
  the ARM child names) through a switch helper in `locals.go`,
  mirroring the Terraform module's lookup table.
- No tags: the peering is an ARM child of the circuit and carries no
  tags argument in the provider schema.

## Inputs

The module receives an `AzureExpressRouteCircuitPeeringIacInput` containing:

- `target.spec.resource_group` / `target.spec.express_route_circuit_name` -- the parent circuit, by NAME (references resolved to literals by the platform)
- `target.spec.peering_type` / `target.spec.vlan_id` -- the peering's ARM identity and VLAN
- `target.spec.primary_peer_address_prefix` / `secondary_peer_address_prefix` -- the /30 session pair
- `target.spec.microsoft_peering_config` / `ipv6` / `route_filter_id` -- Microsoft-peering and IPv6 configuration
- `target.spec.connections` -- Global Reach links, by name
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `express_route_circuit_peering_id` | Full ARM ID -- the far side of another circuit's Global Reach connection |
| `azure_asn` | Microsoft's BGP ASN on the peering |
| `primary_azure_port` / `secondary_azure_port` | Microsoft-edge port identifiers |
| `connection_ids` | Name-keyed Global Reach connection IDs |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
