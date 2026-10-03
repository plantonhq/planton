# AzureExpressRouteGateway Pulumi Module

## Overview

This Pulumi module provisions an ExpressRoute Gateway using the Azure
Classic provider (`pulumi-azure`): a `network.ExpressRouteGateway` plus
one `network.ExpressRouteConnection` per spec entry, each parented
under the gateway and joining a circuit's private peering to the hub.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- The SDK's `ExpressRouteConnectionArgs` still carries the 4.x-era
  deprecated fields (`EnableInternetSecurity`,
  `PrivateLinkFastPathEnabled`); the module uses only the current
  `InternetSecurityEnabled` surface -- the same wire the Terraform
  module speaks at v5.
- `authorization_key` is set only when non-empty (a same-subscription
  circuit needs none), and it is sensitive end to end.
- Connections export a name-keyed ID map (`connection_ids`) so
  downstream tooling references children by the spec's own names.

## Inputs

The module receives an `AzureExpressRouteGatewayIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the gateway's ARM identity (references resolved to literals by the platform)
- `target.spec.virtual_hub_id` -- the hub the gateway deploys into
- `target.spec.scale_units` -- the capacity floor (1-10)
- `target.spec.allow_non_virtual_wan_traffic` -- non-WAN traffic policy
- `target.spec.connections` -- the composed circuit-peering connections
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `express_route_gateway_id` | Full ARM ID of the gateway |
| `express_route_gateway_name` | The gateway's name |
| `connection_ids` | Connection ARM IDs, keyed by name |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
