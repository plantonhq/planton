# AzureVirtualNetworkGatewayConnection Pulumi Module

## Overview

This Pulumi module provisions a gateway connection using the Azure
Classic provider (`pulumi-azure`). It creates a
`network.VirtualNetworkGatewayConnection` -- the tunnel object joining a
virtual network gateway to an on-premises device (IPsec), another
gateway (Vnet2Vnet), or an ExpressRoute circuit.

The connection provisions in seconds; its far-side pairing rules are
spec-validated before the module runs. ARM provisioning success does
NOT mean the tunnel is Connected -- the far side must negotiate.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- **Wire maps in `locals.go`**: spec enums arrive as proto values
  (IPSEC, IKE_V2, DEFAULT) and map to azurerm's exact vocabulary,
  mirroring the Terraform module's `locals.tf` so both engines produce
  identical payloads.
- **`BgpEnabled` over `EnableBgp`**: the classic SDK carries both names
  for the same ARM property; the module uses the modern one.
- **PARITY-EXCEPTION**: the classic SDK models exactly ONE traffic
  selector policy where the provider accepts a list -- the module FAILS
  LOUDLY on more than one, so multi-selector connections deploy via the
  Terraform engine. Silently dropping selectors the user wrote down is
  never acceptable.
- Secrets (`shared_key`, `authorization_key`) are reference-resolved by
  the platform before the module runs and omitted when empty -- Azure
  generates the pre-shared key when absent.

## Inputs

The module receives an `AzureVirtualNetworkGatewayConnectionIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the connection's ARM identity (references resolved to literals by the platform)
- `target.spec.type` -- IPSEC / VNET_TO_VNET / EXPRESS_ROUTE, deciding the required far side
- `target.spec.virtual_network_gateway_id` plus the type's far-side reference
- `target.spec.shared_key` / `target.spec.authorization_key` -- sensitive, omitted when empty
- `target.spec.ipsec_policy` / `target.spec.bgp_enabled` / `target.spec.custom_bgp_addresses` -- tunnel parameters
- `target.spec.egress_nat_rule_ids` / `target.spec.ingress_nat_rule_ids` -- gateway NAT rules this tunnel opts into
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `connection_id` | Full ARM ID of the connection |
| `connection_name` | The connection's name as deployed |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
