# AwsElasticIp — Terraform Module

This directory contains the Terraform module that provisions an AWS Elastic IP.

## Structure

```
tf/
├── main.tf       # aws_eip resource
├── variables.tf  # metadata and spec inputs
├── outputs.tf    # allocation_id, public_ip, arn, public_dns
├── locals.tf     # Tag merging, null coalescing for optional fields
├── provider.tf   # AWS provider configuration
```

## Usage

```hcl
module "eip" {
  source = "./path/to/module"

  metadata = {
    name = "my-eip"
    org  = "acme"
    env  = "prod"
    id   = "my-eip-prod"
  }

  spec = {}
}
```

## Outputs

| Name | Description |
|------|-------------|
| `allocation_id` | EIP allocation ID |
| `public_ip` | Public IPv4 address |
| `arn` | EIP ARN |
| `public_dns` | Public DNS hostname |

## Feature Parity

This Terraform module has full feature parity with the Pulumi module:
- VPC domain (hardcoded)
- BYOIP pool and specific address allocation
- Network border group for Local/Wavelength zones
- Tag merging from metadata
