# AzureExpressRoutePort Pulumi Module

## Overview

This Pulumi module provisions an ExpressRoute Port using the Azure
Classic provider (`pulumi-azure`). It creates a
`network.ExpressRoutePort` plus one
`network.ExpressRoutePortAuthorization` child per spec `authorizations`
entry, parented to the port -- the keys other subscriptions redeem to
build circuits on this port's capacity.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- The SDK types the fixed physical pair as separate `link1`/`link2`
  blocks (`ExpressRoutePortLink1Args`/`ExpressRoutePortLink2Args`), so
  the module carries twin builders that differ only in type.
- The spec's encapsulation/billing/cipher/identity enums map to ARM
  wire values through explicit switch helpers in `locals.go` -- the
  same name-to-wire tables the Terraform module carries in its locals.
  Unset optional enums apply ARM's defaults (MeteredData, GcmAes128),
  mirroring the Terraform variable handling.
- ARM-generated authorization keys are exported through
  `pulumi.ToSecret`, matching the Terraform module's `sensitive = true`
  output.
- The per-link facility facts (router, interface, patch panel, rack)
  export through `Elem()` so an unpopulated fact reads as "" -- the
  Terraform module's `try(..., "")` twin.

## Inputs

The module receives an `AzureExpressRoutePortIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the port's ARM identity (references resolved to literals by the platform)
- `target.spec.peering_location` + `bandwidth_in_gbps` + `encapsulation` -- the physical facts (all ForceNew)
- `target.spec.identity` -- managed identity (USER_ASSIGNED required for MACsec)
- `target.spec.link1` / `link2` -- manipulate the fixed pair (admin state, MACsec)
- `target.spec.authorizations` -- keys to issue, by name
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `express_route_port_id` | Full ARM ID -- what Direct circuits reference |
| `express_route_port_name` | The port's name |
| `guid` / `ethertype` / `mtu` | Port-level physical facts |
| `system_assigned_identity_principal_id` | Populated when the identity type includes SYSTEM_ASSIGNED |
| `link{1,2}_*` | The per-link letter-of-authorization facts |
| `authorization_keys` | Name-keyed generated keys (secret) |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
