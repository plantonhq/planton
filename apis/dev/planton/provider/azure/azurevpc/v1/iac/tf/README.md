# Azure VPC - Terraform Module

This Terraform module provisions an Azure Virtual Network (VNet) with subnets, optional NAT Gateway, and Private DNS zone links.

## Usage

### Basic Configuration

```hcl
module "vpc" {
  source = "./path/to/module"

  metadata = {
    name = "prod-vpc"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    address_space_cidr = "10.0.0.0/16"
    nodes_subnet_cidr  = "10.0.0.0/18"
  }
}
```

### With NAT Gateway

```hcl
module "vpc_with_nat" {
  source = "./path/to/module"

  metadata = {
    name = "prod-vpc"
    env  = "production"
  }

  spec = {
    address_space_cidr     = "10.0.0.0/16"
    nodes_subnet_cidr      = "10.0.0.0/18"
    is_nat_gateway_enabled = true
    tags = {
      "cost-center" = "infrastructure"
    }
  }
}
```

## Inputs

- `metadata.name` (required): Base name for resources
- `spec.address_space_cidr` (required): VNet CIDR block
- `spec.nodes_subnet_cidr` (required): Subnet CIDR block
- `spec.is_nat_gateway_enabled` (optional): Enable NAT Gateway, default false
- `spec.dns_private_zone_links` (optional): Private DNS zone IDs
- `spec.tags` (optional): Additional tags
- `location` (optional): Azure region, default "eastus"

## Outputs

- `vnet_id`: Virtual Network resource ID
- `nodes_subnet_id`: Subnet resource ID
- `resource_group_name`: Resource group name
- `location`: Deployment region
- `nat_gateway_id`: NAT Gateway ID (if enabled)

## NAT Gateway

Enables outbound internet connectivity for private workloads, solving SNAT port exhaustion. See [Research Doc](../docs/README.md) for details.

