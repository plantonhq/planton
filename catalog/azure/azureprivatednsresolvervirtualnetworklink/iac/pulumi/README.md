# AzurePrivateDnsResolverVirtualNetworkLink Pulumi Module

## Overview

Creates the virtual network link that makes a DNS forwarding ruleset take effect in ONE virtual network, on the classic Pulumi Azure SDK (`pulumi-azure/sdk/v6`), wire-identical to the Terraform module.

## Resources Created

- `privatedns.ResolverVirtualNetworkLink` -- the link (a child of the forwarding ruleset)

## Outputs

- `virtual_network_link_id` -- the link's ARM resource ID
- `virtual_network_link_name` -- the link's name

## Behavior Notes

- **The linked network must be in the ruleset's region** but does NOT need to be peered with the resolver's network; cross-subscription links are allowed.
- **Everything except `metadata` is create-only** -- changing the name, ruleset, or network replaces the link.
- **The link carries NO tags** -- ARM gives it a free-form `metadata` map instead.
- **Links are free at rest.**

## Engine Parity

ZERO parity exceptions: the classic SDK v6.38.0 carries the complete v5 argument surface, verified field-by-field at planning.
