# AzureVirtualNetworkGateway Pulumi Module

## Overview

This Pulumi module provisions an Azure virtual network gateway using the
Azure Classic provider (`pulumi-azure`). It creates a
`network.VirtualNetworkGateway` -- the managed appliance terminating
site-to-site VPN, point-to-site VPN, VNet-to-VNet, or ExpressRoute
connectivity -- plus one `network.VirtualNetworkGatewayNatRule` per
composed NAT rule (parented to the gateway).

The gateway is one third of the site-to-site story: it lives in the
referenced "GatewaySubnet" (an ARM name contract) and binds referenced
AzurePublicIp addresses; AzureLocalNetworkGateway describes each
on-premises site and AzureVirtualNetworkGatewayConnection ties a site to
this gateway. Gateways provision in 25-45 minutes and delete in 10-20,
so identity-level changes are expensive -- design to avoid replacement.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- **Wire maps in `locals.go`**: spec enums arrive as proto values
  (VPN_GW_1_AZ, EXPRESS_ROUTE, EGRESS_SNAT) and map to azurerm's exact
  vocabulary, mirroring the Terraform module's `locals.tf` so both
  engines produce identical payloads.
- **`BgpEnabled` over `EnableBgp`**: the classic SDK carries both names
  for the same ARM property; the module uses the modern one.
- **PARITY-EXCEPTION**: the classic SDK does not expose the ER_GW_SCALE
  autoscale bounds (`minimum_scale_unit`/`maximum_scale_unit`) -- the
  module FAILS LOUDLY when a manifest sets them, so autoscaling
  ExpressRoute gateways deploy via the Terraform engine. Silently
  dropping an autoscale contract the user wrote down is never
  acceptable.
- Omission semantics match Terraform: `dns_forwarding_enabled` sends
  only when true (ARM rejects it on unsupporting SKUs);
  `ip_sec_replay_protection_enabled` sends only when the user takes an
  explicit position (the provider default matches the spec default);
  an empty ip configuration name becomes `vnetGatewayConfig`.

## Inputs

The module receives an `AzureVirtualNetworkGatewayIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the gateway's ARM identity (references resolved to literals by the platform)
- `target.spec.type` / `target.spec.vpn_type` / `target.spec.sku` / `target.spec.generation` -- the gateway shape (pairing rules spec-validated)
- `target.spec.ip_configurations` -- 1-3 bindings of the GatewaySubnet + (VPN) public IPs
- `target.spec.bgp_enabled` / `target.spec.bgp_settings` -- the BGP speaker
- `target.spec.vpn_client_configuration` / `target.spec.policy_groups` -- point-to-site
- `target.spec.nat_rules` -- composed NAT rules
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `virtual_network_gateway_id` | Full ARM ID of the gateway -- the join key connections attach through |
| `virtual_network_gateway_name` | The gateway's name as deployed |
| `nat_rule_ids` | Map of NAT rule name to ARM id -- connections opt in via their egress/ingress lists |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
