# AzureSubnet Terraform Module

Terraform implementation for the AzureSubnet deployment component.

## Resources Created

- `azurerm_subnet` -- Subnet within an existing Virtual Network

## Usage

```hcl
module "subnet" {
  source = "./iac/tf"

  metadata = {
    name = "app-subnet"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    resource_group = "prod-network-rg"
    vnet_id        = "/subscriptions/.../virtualNetworks/prod-vnet"
    name           = "app-subnet"
    address_prefix = "10.0.1.0/24"
    service_endpoints = ["Microsoft.Sql", "Microsoft.Storage"]
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata | object | yes |
| spec | Subnet specification | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| subnet_id | Azure Resource Manager ID |
| subnet_name | Name of the subnet |
| address_prefix | IPv4 CIDR block |

## Key Patterns

- VNet name is extracted from the ARM resource ID in `locals.tf`
- Address prefix is wrapped in a list: `[var.spec.address_prefix]`
- Delegation uses `dynamic` block (present only when `var.spec.delegation` is non-null)
