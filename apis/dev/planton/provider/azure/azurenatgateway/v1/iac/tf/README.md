# Azure NAT Gateway - Terraform Module

This Terraform module provisions an Azure NAT Gateway with dynamic SNAT, eliminating port exhaustion for private workloads.

## Usage

### Basic Configuration

```hcl
module "nat_gateway" {
  source = "./path/to/module"

  metadata = {
    name = "prod-nat"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    subnet_id = "/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{vnet}/subnets/{subnet}"
  }
}
```

### Production with IP Prefix

```hcl
module "nat_gateway_prod" {
  source = "./path/to/module"

  metadata = {
    name = "prod-nat"
    env  = "production"
  }

  spec = {
    subnet_id               = var.subnet_id
    idle_timeout_minutes    = 10
    public_ip_prefix_length = 28  # 16 IPs, 1M+ ports
    tags = {
      "cost-center" = "infrastructure"
    }
  }
}
```

## Inputs

- `metadata.name` (required): Base name for resources
- `spec.subnet_id` (required): Subnet resource ID
- `spec.idle_timeout_minutes` (optional): Timeout 4-120, default 4
- `spec.public_ip_prefix_length` (optional): Prefix /28-/31
- `spec.tags` (optional): Additional resource tags

## Outputs

- `nat_gateway_id`: NAT Gateway resource ID
- `public_ip_addresses`: List of allocated IPs (if using individual IP)
- `public_ip_prefix_id`: Prefix ID (if using prefix)

## Important Notes

- Use Public IP Prefix (28-31) for production scale
- Single IP provides 64,512 ports (sufficient for many workloads)
- /28 prefix provides 16 IPs = 1,032,192 ports
- NAT Gateway takes precedence over all other outbound methods
- Cost: ~$0.045/hour + $0.045/GB processed

See full documentation: [Research Doc](../docs/README.md)

