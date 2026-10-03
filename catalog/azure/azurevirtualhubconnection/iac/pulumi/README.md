# AzureVirtualHubConnection Pulumi Module

## Overview

This Pulumi module provisions a Virtual Hub Connection using the Azure
Classic provider (`pulumi-azure`). It creates a single
`network.VirtualHubConnection` -- the attachment joining one spoke
virtual network to a Virtual WAN hub, with the routing block where WAN
topologies are expressed.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- The connection carries no tags and no resource group of its own (ARM
  addresses it as a child of the hub; the provider has no tags
  surface), so the module skips the family's usual tag-merging locals.
- The spec's optional routing fields apply ARM's defaults when unset
  (override criteria Contains, static-route propagation ON) through
  nil-handling helpers in `locals.go`, mirroring the Terraform module's
  null handling.
- All ID fields are StringValueOrRef -- the platform resolves valueFrom
  references (the hub's `route_table_ids.<name>` map outputs
  especially) to literals before the module runs.

## Inputs

The module receives an `AzureVirtualHubConnectionIacInput` containing:

- `target.spec.name` -- the connection's name (2-80 chars, the provider's regex)
- `target.spec.virtual_hub_id` / `target.spec.remote_virtual_network_id` -- the two sides of the attachment
- `target.spec.internet_security_enabled` -- default-route advertisement
- `target.spec.routing` -- association, propagation, static routes
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `virtual_hub_connection_id` | Full ARM ID -- what a hub BGP peering references |
| `virtual_hub_connection_name` | The connection's name |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
