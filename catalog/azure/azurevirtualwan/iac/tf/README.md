# AzureVirtualWan Terraform Module

## Overview

This Terraform module provisions a Virtual WAN using the `azurerm`
provider. It creates a single `azurerm_virtual_wan` -- the free,
lightweight umbrella object of Azure's managed hub-and-spoke
networking. Virtual hubs (and the gateways on them) are separate
resources that reference this WAN's ID.

## Resources Created

- `azurerm_virtual_wan.main` -- the WAN

## Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Planton metadata (name, org, env) |
| `spec` | object | Virtual WAN specification |

Key spec fields:

| Field | Required | Description |
|-------|----------|-------------|
| `region` / `resource_group` / `name` | yes | The WAN's ARM identity; all ForceNew |
| `disable_vpn_encryption` | no | Off by default (traffic encrypted) |
| `allow_branch_to_branch_traffic` | no | Defaults to true (ARM's default) |
| `office365_local_breakout_category` | no | NONE (default) / OPTIMIZE / OPTIMIZE_AND_ALLOW / ALL |
| `type` | no | "Standard" (default, full mesh) or "Basic" (legacy, upgrade-only) |
| `tags` | no | User tags, merged over metadata-derived tags |

## Outputs

| Output | Description |
|--------|-------------|
| `virtual_wan_id` | Full ARM ID -- what virtual hubs reference |
| `virtual_wan_name` | The WAN's name |

## Usage

```hcl
module "global_wan" {
  source = "./iac/tf"

  metadata = { name = "global-wan", org = "mycompany", env = "production" }

  spec = {
    region         = "eastus"
    resource_group = "network-rg"
    name           = "global-wan"
  }
}
```

## Behavior Notes

- The WAN object itself is free; hubs and gateways carry the cost.
- ARM refuses to delete a WAN that still has hubs -- destroy hubs
  first.
- A Basic WAN upgrades to Standard in place; Standard never
  downgrades.

## Required Permissions

Least-privilege runner permissions for this kind are declared in [`../permissions.yaml`](../permissions.yaml).
