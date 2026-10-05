# AzurePointToSiteVpnGateway Pulumi Module

## Overview

This Pulumi module provisions a Point-to-Site VPN Gateway using the
Azure Classic provider (`pulumi-azure`). It creates a single
`network.PointToPointVpnGateway` -- the managed receiver inside a
Virtual WAN hub that individual devices dial into.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- The SDK's resource lives at the legacy, misnamed
  `PointToPointVpnGateway` token but creates the SAME ARM
  `p2sVpnGateways` object as `azurerm_point_to_site_vpn_gateway` --
  do not "fix" the name (also recorded in the import catalog).
- `virtual_hub_id`, `vpn_server_configuration_id`, and the route
  block's table/map references are StringValueOrRef -- the platform
  resolves valueFrom references to literals before the module runs.
- The spec's unset `scale_unit` renders as an explicit 1 (the
  provider REQUIRES a value and has no default), mirroring the
  Terraform module.
- Route-map IDs and `dns_servers` are wired only when configured
  (never sent as empty), mirroring the Terraform module's null
  handling.

## Inputs

The module receives an `AzurePointToSiteVpnGatewayIacInput` containing:

- `target.spec.virtual_hub_id` -- the hub the gateway deploys into (one P2S gateway per hub)
- `target.spec.vpn_server_configuration_id` -- the authentication policy
- `target.spec.connection_configurations` -- named client address pools with optional per-pool routing
- `target.spec.scale_unit` / `routing_preference_internet_enabled` / `dns_servers`
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `point_to_site_vpn_gateway_id` | Full ARM ID of the gateway |
| `point_to_site_vpn_gateway_name` | The gateway's name |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
