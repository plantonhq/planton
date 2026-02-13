# AzureResourceGroup Terraform Module

## Overview

This Terraform module provisions an Azure Resource Group using the `azurerm` provider.
It creates a single `azurerm_resource_group` with OpenMCF metadata tags.

## Resources Created

- `azurerm_resource_group.main` -- the Azure Resource Group

## Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | OpenMCF metadata (name, org, env) |
| `spec` | object | Resource group specification (name, region) |

## Outputs

| Output | Description |
|--------|-------------|
| `resource_group_id` | Azure Resource Manager ID |
| `resource_group_name` | Name of the created resource group |
| `region` | Azure region |

## Usage

```hcl
module "resource_group" {
  source = "./iac/tf"

  metadata = {
    name = "platform-rg"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    name   = "prod-platform-rg"
    region = "eastus"
  }
}
```
