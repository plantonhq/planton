# AzurePrivateDnsZone Terraform Module

## Overview

This Terraform module creates an Azure Private DNS Zone with a Virtual Network link.

## Resources Created

- `azurerm_private_dns_zone.zone` -- The private DNS zone (global, no region)
- `azurerm_private_dns_zone_virtual_network_link.vnet_link` -- Links the zone to a VNet

## Variables

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `metadata` | object | Yes | Resource metadata (name, id, org, env, labels, tags) |
| `spec` | object | Yes | Zone spec (resource_group, name, vnet_id, registration_enabled) |

## Outputs

| Output | Description |
|--------|-------------|
| `zone_id` | Azure Resource Manager ID of the zone |
| `zone_name` | Name of the private DNS zone |

## Usage

```hcl
module "pg_private_dns" {
  source = "./iac/tf"

  metadata = {
    name = "pg-private-dns"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    resource_group       = "prod-network-rg"
    name                 = "privatelink.postgres.database.azure.com"
    vnet_id              = "/subscriptions/.../virtualNetworks/prod-vnet"
    registration_enabled = false
  }
}
```

## Provider

Requires `hashicorp/azurerm` ~> 4.0.
