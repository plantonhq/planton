# DigitalOcean VPC - Terraform Module

This Terraform module provisions and manages DigitalOcean Virtual Private Cloud (VPC) networks based on the Planton DigitalOceanVpc specification.

## Overview

The module creates fully-configured DigitalOcean VPCs with support for:
- Auto-generated CIDR blocks (80% use case)
- Explicit CIDR control (20% use case)  
- Regional isolation
- Optional descriptions
- Immutability-aware lifecycle management

## Prerequisites

1. **Terraform**: Version 1.5 or higher
2. **DigitalOcean Account**: Active account with API access
3. **DigitalOcean API Token**: Personal access token with read/write permissions

## Installation

```bash
terraform init
```

This downloads the DigitalOcean provider (~2.0).

### Set Required Variables

```bash
export TF_VAR_digitalocean_token="your-do-api-token-here"
```

**Security Note:** Never commit API tokens to version control.

## Usage

### Example 1: Minimal VPC (Auto-Generated CIDR)

**80% Use Case:** Let DigitalOcean auto-generate a non-conflicting /20 CIDR block.

```hcl
module "dev_vpc" {
  source = "./path/to/module"

  metadata = {
    name = "dev-vpc"
    env  = "development"
  }

  spec = {
    region = "nyc3"
    # ip_range_cidr is intentionally omitted
    # DigitalOcean will auto-generate a /20 block
  }

  digitalocean_token = var.digitalocean_token
}
```

**Result:** DigitalOcean auto-generates a CIDR like `10.116.0.0/20` (4,096 IPs).

### Example 2: Production VPC with Explicit /16 CIDR

**20% Use Case:** Explicit IP planning for production environments.

```hcl
module "prod_vpc" {
  source = "./path/to/module"

  metadata = {
    name = "prod-vpc"
    env  = "production"
    tags = ["criticality:high", "team:platform"]
  }

  spec = {
    description   = "Main production VPC for all services"
    region        = "nyc1"
    ip_range_cidr = "10.101.0.0/16"  # 65,536 IPs for growth
  }

  digitalocean_token = var.digitalocean_token
}
```

### Example 3: Multi-Environment Setup

```hcl
# Development VPC (auto-generated)
module "dev_vpc" {
  source = "./path/to/module"

  metadata = { name = "dev-vpc" }
  spec     = { region = "nyc3" }
  
  digitalocean_token = var.digitalocean_token
}

# Staging VPC (explicit /20)
module "staging_vpc" {
  source = "./path/to/module"

  metadata = { name = "staging-vpc" }
  spec = {
    region        = "nyc3"
    ip_range_cidr = "10.100.16.0/20"
    description   = "Staging environment - production-like config"
  }
  
  digitalocean_token = var.digitalocean_token
}

# Production VPC (explicit /16)
module "prod_vpc" {
  source = "./path/to/module"

  metadata = { name = "prod-vpc" }
  spec = {
    region        = "nyc1"
    ip_range_cidr = "10.101.0.0/16"
    description   = "Production VPC - immutable after creation"
  }
  
  digitalocean_token = var.digitalocean_token
}
```

## Variables

### Required Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Resource metadata (name, env, tags, etc.) |
| `spec` | object | VPC specification (see spec variables) |
| `digitalocean_token` | string | DigitalOcean API token (sensitive) |

### Spec Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `description` | string | No | "" | Human-readable VPC description |
| `region` | string | Yes | - | DigitalOcean region (e.g., "nyc3", "sfo3") |
| `ip_range_cidr` | string | No | "" | CIDR block (/16, /20, or /24). Auto-generated if omitted. |
| `is_default_for_region` | bool | No | false | Set as regional default VPC |

## Outputs

| Output | Description |
|--------|-------------|
| `vpc_id` | VPC UUID (use this in cluster/database vpc_uuid fields) |
| `vpc_urn` | DigitalOcean URN |
| `ip_range` | Actual IP range (useful for auto-generated CIDRs) |
| `is_default` | Whether VPC is regional default |
| `created_at` | Creation timestamp |
| `region` | VPC region |
| `vpc_name` | VPC name |

### Accessing Outputs

```bash
# Get VPC ID for use in other resources
terraform output vpc_id

# Get auto-generated IP range
terraform output ip_range

# All outputs
terraform output
```

## Deployment Workflow

### 1. Plan

```bash
terraform plan -out=tfplan
```

Review:
- VPC name and region are correct
- IP range is either specified or will be auto-generated
- No CIDR conflicts with existing VPCs

### 2. Apply

```bash
terraform apply tfplan
```

**Deployment time:** ~30 seconds

### 3. Verify

```bash
# Get VPC ID
VPC_ID=$(terraform output -raw vpc_id)
echo "VPC ID: $VPC_ID"

# Get IP range (especially useful for auto-generated)
IP_RANGE=$(terraform output -raw ip_range)
echo "IP Range: $IP_RANGE"

# Verify via DigitalOcean CLI
doctl vpcs get $VPC_ID
```

### 4. Use in Dependent Resources

```hcl
# Create DOKS cluster in this VPC
module "k8s_cluster" {
  source = "../digitaloceankubernetescluster/v1/iac/tf"
  
  spec = {
    vpc = {
      value = module.prod_vpc.vpc_id
    }
    # ... other config
  }
}
```

## Common Operations

### Get Auto-Generated IP Range

For VPCs with auto-generated CIDR, retrieve the assigned range:

```bash
terraform output ip_range
```

### Update VPC Description

**Note:** IP range is immutable and in `ignore_changes`, but description can be updated.

```hcl
spec = {
  description = "Updated description"
  # ... other config
}
```

```bash
terraform apply
```

### Destroy VPC

**Warning:** VPC deletion fails if resources (Droplets, clusters, databases) still exist in it.

```bash
# List VPC resources
VPC_ID=$(terraform output -raw vpc_id)
doctl vpcs resources get $VPC_ID

# Delete all resources first, then:
terraform destroy
```

## 80/20 Principle in Practice

### 80% Use Case: Auto-Generated CIDR

For most dev/test environments:

```hcl
spec = {
  region = "nyc3"
  # Omit ip_range_cidr completely
}
```

**Benefits:**
- Zero IP planning overhead
- DigitalOcean ensures no conflicts
- Fast deployment
- Perfect for throwaway environments

### 20% Use Case: Explicit CIDR Planning

For production with IPAM requirements:

```hcl
spec = {
  region        = "nyc1"
  ip_range_cidr = "10.101.0.0/16"
  description   = "Production VPC - documented in IPAM"
}
```

**Benefits:**
- Full control over IP allocation
- Non-overlapping ranges for VPC peering
- Integrates with corporate IPAM
- Supports multi-region strategies

## State Management

### Remote State (Recommended)

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "digitalocean/vpcs/prod-vpc.tfstate"
    region = "us-east-1"
  }
}
```

### Terraform Cloud

```hcl
terraform {
  backend "remote" {
    organization = "my-org"
    workspaces {
      name = "digitalocean-prod-vpc"
    }
  }
}
```

## CIDR Planning Guide

### T-Shirt Sizing

| Environment | Recommended CIDR | IPs Available | Use Case |
|-------------|------------------|---------------|----------|
| **Dev/Test** | Auto-generated /20 | ~4,000 | Quick experiments |
| **Small Prod** | x.x.0.0/20 | ~4,000 | Small production |
| **Medium Prod** | x.x.0.0/18 | ~16,000 | Growing production |
| **Large Prod** | x.x.0.0/16 | ~65,000 | DOKS clusters + databases |

### Reserved Ranges (Avoid)

DigitalOcean reserves these ranges:
- `10.244.0.0/16`
- `10.245.0.0/16`
- `10.246.0.0/24`
- `10.229.0.0/16`
- `10.10.0.0/16` in nyc1 region

### Non-Overlapping Strategy

```hcl
# Development: 10.100.0.0/20
# Staging:     10.100.16.0/20
# Production:  10.101.0.0/16

# No overlap → VPC peering is possible
```

## Best Practices

### 1. VPC-First Deployment

**Always create VPCs before:**
- DOKS clusters (cannot be migrated)
- Load balancers (cannot be migrated)
- Managed databases (painful to migrate)

### 2. Plan for Immutability

**VPC IP ranges cannot be changed.** Over-provision:
- Dev: /20 is usually sufficient
- Production: Use /16 to avoid future migrations

### 3. Use Auto-Generation for Dev

Don't waste time planning dev/test IP ranges:

```hcl
spec = {
  region = "nyc3"
  # Let DigitalOcean handle it
}
```

### 4. Explicit Planning for Production

Production VPCs need explicit CIDRs:
- Document in IPAM
- Ensure no overlap with other environments
- Plan for VPC peering

### 5. Tag Everything

```hcl
metadata = {
  name = "prod-vpc"
  tags = ["env:production", "team:platform", "managed:terraform"]
}
```

## Security

1. **API Token Security**
   - Store in environment variables
   - Use Terraform Cloud for team access
   - Rotate regularly

2. **Private Networking**
   - All VPC traffic is private by default
   - Free internal traffic
   - No public exposure

3. **Network Isolation**
   - Separate VPCs per environment
   - Use firewall rules within VPCs
   - Implement least-privilege access

## Troubleshooting

### "CIDR block overlaps"

**Symptom:** VPC creation fails with overlap error

**Solution:**
```bash
# List existing VPCs
doctl vpcs list --format Name,IPRange

# Choose non-overlapping CIDR
```

### "Invalid CIDR format"

**Valid formats:**
- `/16` - 10.100.0.0/16
- `/20` - 10.100.16.0/20
- `/24` - 10.100.16.0/24

**Invalid:**
- `/8`, `/12`, `/18`, `/22`, `/28`

### "Cannot delete VPC"

**Symptom:** "VPC has resources attached"

**Solution:**
```bash
# List attached resources
doctl vpcs resources get $(terraform output -raw vpc_id)

# Delete all resources first
# Then destroy VPC
terraform destroy
```

### "Reserved IP range"

**Symptom:** Creation fails with reserved range error

**Solution:** Avoid these ranges:
- 10.244.0.0/16
- 10.245.0.0/16
- 10.246.0.0/24
- 10.229.0.0/16

Use 10.100.0.0/16 - 10.199.0.0/16 instead.

## Cost Information

**VPCs are free on DigitalOcean.**

**Benefits:**
- No monthly VPC charges
- Free internal traffic
- Free VPC peering (same datacenter)

## Module Maintenance

### Update Provider

```hcl
required_providers {
  digitalocean = {
    version = "~> 2.1"  # Update version
  }
}
```

```bash
terraform init -upgrade
```

## Reference

- [DigitalOcean Terraform Provider](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/vpc)
- [DigitalOcean VPC Documentation](https://docs.digitalocean.com/products/networking/vpc/)
- [RFC1918 Private Address Space](https://tools.ietf.org/html/rfc1918)
- [Planton Documentation](https://docs.planton.dev/)

---

**Version:** 1.0.0  
**Last Updated:** 2025-11-16

