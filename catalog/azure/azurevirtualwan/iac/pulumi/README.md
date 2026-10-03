# AzureVirtualWan Pulumi Module

## Overview

This Pulumi module provisions a Virtual WAN using the Azure Classic
provider (`pulumi-azure`). It creates a single `network.VirtualWan` --
the free, lightweight umbrella object of Azure's managed hub-and-spoke
networking, which virtual hubs and their gateways reference.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- The spec's optional fields apply ARM's defaults when unset
  (branch-to-branch ON, Office 365 breakout None, type "Standard")
  through nil-handling helpers in `locals.go`, mirroring the Terraform
  module's null handling.
- The Office 365 breakout enum maps to ARM wire values through an
  explicit switch helper -- the same name-to-wire table the Terraform
  module carries in its locals.

## Inputs

The module receives an `AzureVirtualWanIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the WAN's ARM identity (references resolved to literals by the platform)
- `target.spec.disable_vpn_encryption` / `allow_branch_to_branch_traffic` -- transit policy
- `target.spec.office365_local_breakout_category` -- local breakout scope
- `target.spec.type` -- "Standard" (default) or "Basic"
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `virtual_wan_id` | Full ARM ID -- what virtual hubs reference |
| `virtual_wan_name` | The WAN's name |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
