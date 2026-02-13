# AzureNetworkSecurityGroup Terraform Module

Terraform implementation for the AzureNetworkSecurityGroup deployment component.

## Resources Created

- `azurerm_network_security_group` -- Network Security Group
- `azurerm_network_security_rule` -- One per security rule (separate resources)

## Usage

```hcl
module "nsg" {
  source = "./iac/tf"

  metadata = {
    name = "web-tier-nsg"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region         = "eastus"
    resource_group = "prod-network-rg"
    name           = "web-tier-nsg"
    security_rules = [
      {
        name                   = "allow-https"
        priority               = 100
        direction              = "Inbound"
        access                 = "Allow"
        protocol               = "Tcp"
        destination_port_range = "443"
      },
      {
        name                   = "deny-all-inbound"
        priority               = 4096
        direction              = "Inbound"
        access                 = "Deny"
        protocol               = "*"
        destination_port_range = "*"
      }
    ]
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata | object | yes |
| spec | NSG specification | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| nsg_id | Azure Resource Manager ID |
| nsg_name | Name of the Network Security Group |
