# AzureVpnSite Pulumi Module

## Overview

This Pulumi module provisions a VPN Site using the Azure Classic
provider (`pulumi-azure`). It creates a single `network.VpnSite` -- the
Virtual WAN address-book entry for one branch location, with its links,
reachable address space, device metadata, and O365 breakout policy.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- ARM assigns each link an ID; the module republishes them keyed by the
  link's NAME (`link_ids`) so connections reference
  `status.outputs.link_ids.<link-name>` -- the family's name-keyed map
  convention.
- Empty optional strings are omitted (never sent as "") -- the
  provider validates configured values as non-empty, mirroring the
  Terraform module's null handling.
- `virtual_wan_id` and `resource_group` are StringValueOrRef -- the
  platform resolves valueFrom references to literals before the module
  runs.

## Inputs

The module receives an `AzureVpnSiteIacInput` containing:

- `target.spec.name` -- the site's name (the provider's character rule)
- `target.spec.virtual_wan_id` -- the WAN the site belongs to
- `target.spec.address_cidrs` / `target.spec.links` -- the routing source and connectable endpoints
- `target.spec.o365_policy` -- O365 breakout categories
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `vpn_site_id` | Full ARM ID -- what a connection's `remote_vpn_site_id` references |
| `vpn_site_name` | The site's name |
| `link_ids` | Each link's ARM ID keyed by link name -- what a connection's tunnels pin to |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
