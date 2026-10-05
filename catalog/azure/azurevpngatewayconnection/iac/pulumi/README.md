# AzureVpnGatewayConnection Pulumi Module

## Overview

This Pulumi module provisions a VPN Gateway Connection using the Azure
Classic provider (`pulumi-azure`). It creates a single
`network.VpnGatewayConnection` -- the tunnel bundle joining one branch
(a VPN Site) to a Virtual WAN hub's VPN gateway, with per-tunnel IPsec,
BGP, and NAT choices.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- The connection carries no tags, region, or resource group of its own
  (ARM addresses it as a child of the gateway; the provider has no
  tags surface), so the module skips the family's usual tag-merging
  locals.
- ARM's per-link defaults are rendered explicitly (bandwidth 10,
  protocol "IKEv2", mode "Default") through nil-handling helpers in
  `locals.go`, mirroring the Terraform module's null handling;
  `dpd_timeout_seconds` and `shared_key` are omitted when unset.
- Every ID field is StringValueOrRef -- the platform resolves valueFrom
  references (the site's `link_ids.<name>` and the gateway's
  `nat_rule_ids.<name>` map outputs especially) to literals before the
  module runs.

## Inputs

The module receives an `AzureVpnGatewayConnectionIacInput` containing:

- `target.spec.name` -- the connection's name
- `target.spec.vpn_gateway_id` / `target.spec.remote_vpn_site_id` -- the two sides
- `target.spec.vpn_links` -- one tunnel per site link, each with its own parameters
- `target.spec.routing` -- hub route-table association and propagation
- `target.spec.traffic_selector_policies` -- optional CIDR restrictions
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `connection_id` | Full ARM ID of the connection |
| `connection_name` | The connection's name |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
