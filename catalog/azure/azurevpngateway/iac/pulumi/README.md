# AzureVpnGateway Pulumi Module

## Overview

This Pulumi module provisions a Virtual WAN VPN Gateway using the Azure
Classic provider (`pulumi-azure`). It creates a `network.VpnGateway`
(the managed site-to-site terminator inside a virtual hub) plus one NAT
rule child per spec entry.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- The NAT rule child's SDK constructor is `NewVnpGatewayNatRule` -- the
  resource lives at a legacy, TYPO'D "VnpGatewayNatRule" token but
  creates the same ARM object as `azurerm_vpn_gateway_nat_rule`; do
  not "fix" the name.
- ARM's defaults are rendered explicitly (routing preference
  "Microsoft Network", scale unit 1, NAT mode/type
  "EgressSnat"/"Static") through nil-handling helpers in `locals.go`,
  mirroring the Terraform module's null handling.
- ARM assigns the instance public/private IPs and the BGP facts at
  creation; the module republishes them (`public_ip_addresses`,
  `bgp_asn`) -- what branch device configuration consumes.
- `virtual_hub_id` and `resource_group` are StringValueOrRef -- the
  platform resolves valueFrom references to literals before the module
  runs.

## Inputs

The module receives an `AzureVpnGatewayIacInput` containing:

- `target.spec.name` -- the gateway's name
- `target.spec.virtual_hub_id` -- the hub to deploy into (one VPN gateway per hub)
- `target.spec.scale_unit` / `target.spec.routing_preference` -- capacity and egress path
- `target.spec.bgp_settings` -- ASN, peer weight, per-instance custom APIPA IPs
- `target.spec.nat_rules` -- the composed translation children
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `vpn_gateway_id` | Full ARM ID -- what a connection's `vpn_gateway_id` references |
| `vpn_gateway_name` | The gateway's name |
| `bgp_asn` | The gateway's ASN (65515 on today's Virtual WAN) |
| `public_ip_addresses` | Each instance's public IPv4 -- what branch devices dial |
| `private_ip_addresses` | Each instance's private IPv4 |
| `nat_rule_ids` | Each NAT rule's ARM ID keyed by rule name |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
