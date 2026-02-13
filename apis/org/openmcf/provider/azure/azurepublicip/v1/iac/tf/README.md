# AzurePublicIp Terraform Module

Terraform implementation for the AzurePublicIp deployment component.

## Resources Created

- `azurerm_public_ip` -- Standard SKU, Static allocation Public IP Address

## Usage

```hcl
module "public_ip" {
  source = "./iac/tf"

  metadata = {
    name = "my-pip"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region         = "eastus"
    resource_group = "my-rg"
    name           = "my-public-ip"
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata | object | yes |
| spec | Public IP specification | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| public_ip_id | Azure Resource Manager ID |
| ip_address | Allocated static IPv4 address |
| fqdn | FQDN (if domain_name_label set) |
| public_ip_name | Name of the Public IP |
