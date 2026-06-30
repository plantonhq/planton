# DigitalOcean Spaces Bucket - Terraform Module

## Overview

This Terraform module deploys DigitalOcean Spaces buckets (S3-compatible object storage) using declarative Infrastructure as Code. The module provides type-safe variable definitions, automatic validation, and comprehensive outputs matching the DigitalOceanBucket protobuf specification.

## Features

- **S3-Compatible Storage**: Full S3 API compatibility
- **Access Control**: Private or public-read configurations
- **Versioning**: Object versioning for data protection
- **Built-in Validation**: Terraform validation rules ensure correct configuration
- **Comprehensive Outputs**: All bucket metadata exported
- **State Management**: Terraform tracks resources and detects drift

## Prerequisites

- **Terraform**: Version 1.0 or higher
- **DigitalOcean Account**: With Spaces enabled
- **DigitalOcean API Token**: For infrastructure management
- **Spaces Access Keys**: For S3-compatible object access (optional)

## Installation

This module is part of the Planton monorepo:

```bash
git clone https://github.com/plantonhq/planton.git
cd planton/apis/dev/planton/provider/digitalocean/digitaloceanbucket/v1/iac/tf
```

## Usage

### Basic Example

```hcl
module "app_bucket" {
  source = "./path/to/module"

  metadata = {
    name = "my-app-data"
    labels = {
      environment = "production"
      team        = "backend"
    }
  }

  spec = {
    bucket_name        = "my-app-data"
    region             = "nyc3"
    access_control     = 0  # PRIVATE
    versioning_enabled = false
  }

  digitalocean_token = var.digitalocean_token
}

output "bucket_endpoint" {
  value = module.app_bucket.endpoint
}
```

### Public Bucket for Static Website

```hcl
module "static_site" {
  source = "./path/to/module"

  metadata = {
    name = "company-website"
  }

  spec = {
    bucket_name    = "company-website"
    region         = "nyc3"
    access_control = 1  # PUBLIC_READ
    tags = [
      "website",
      "cdn-enabled",
      "public"
    ]
  }

  digitalocean_token = var.digitalocean_token
}
```

### Versioned Backup Bucket

```hcl
module "backup_bucket" {
  source = "./path/to/module"

  metadata = {
    name = "database-backups"
    labels = {
      environment        = "production"
      data-classification = "critical"
    }
  }

  spec = {
    bucket_name        = "prod-db-backups"
    region             = "sfo3"
    access_control     = 0  # PRIVATE
    versioning_enabled = true
    tags = [
      "backups",
      "versioned",
      "retention-30days"
    ]
  }

  digitalocean_token = var.digitalocean_token
}
```

## Module Inputs

### Required Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Resource metadata (name, labels) |
| `spec` | object | Bucket specification (see below) |
| `digitalocean_token` | string | DigitalOcean API token (sensitive) |

### Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `bucket_name` | string | Yes | - | DNS-compatible bucket name (3-63 chars) |
| `region` | string | Yes | - | DigitalOcean region (`nyc3`, `sfo3`, etc.) |
| `access_control` | number | No | 0 | `0` = PRIVATE, `1` = PUBLIC_READ |
| `versioning_enabled` | bool | No | false | Enable object versioning |
| `tags` | list(string) | No | [] | Bucket tags |

### Optional Variables

| Variable | Type | Description |
|----------|------|-------------|
| `spaces_access_id` | string | Spaces access key ID (for S3 API access) |
| `spaces_secret_key` | string | Spaces secret key (for S3 API access) |

## Module Outputs

| Output | Description |
|--------|-------------|
| `bucket_id` | Unique bucket identifier (`region:bucket-name`) |
| `endpoint` | Bucket FQDN endpoint |
| `bucket_name` | Bucket name |
| `region` | Bucket region |
| `urn` | DigitalOcean URN |
| `bucket_domain_name` | Full bucket domain name |

## Validation Rules

The module includes automatic validation:

1. **Bucket Name Pattern**: Must match `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
2. **Bucket Name Length**: 3-63 characters
3. **Access Control**: Must be 0 (PRIVATE) or 1 (PUBLIC_READ)

## Access Control Values

| Value | Enum Name | ACL String | Public Access |
|-------|-----------|------------|---------------|
| 0 | PRIVATE | `private` | No |
| 1 | PUBLIC_READ | `public-read` | Yes |

## Terraform Workflow

### Initialize

```bash
terraform init
```

### Plan

```bash
export TF_VAR_digitalocean_token="dop_v1_..."
terraform plan
```

### Apply

```bash
terraform apply
```

### Get Outputs

```bash
terraform output bucket_endpoint
terraform output bucket_id
```

### Destroy

```bash
terraform destroy
```

## Best Practices

### Credential Management

**Never commit credentials**. Use one of these methods:

#### Environment Variables

```bash
export TF_VAR_digitalocean_token="your-token"
export TF_VAR_spaces_access_id="your-spaces-key"
export TF_VAR_spaces_secret_key="your-spaces-secret"
terraform apply
```

#### Terraform Cloud

Store sensitive variables in workspace variables marked as sensitive.

#### HashiCorp Vault

```hcl
data "vault_generic_secret" "do_creds" {
  path = "secret/digitalocean/prod"
}

module "bucket" {
  source = "./path/to/module"
  digitalocean_token = data.vault_generic_secret.do_creds.data["token"]
  # ... other config ...
}
```

### State Management

Use remote state for production:

```hcl
terraform {
  backend "s3" {
    bucket   = "terraform-state"
    key      = "spaces/my-bucket/terraform.tfstate"
    region   = "us-east-1"
    endpoint = "nyc3.digitaloceanspaces.com"
    
    skip_credentials_validation = true
    skip_metadata_api_check     = true
    skip_region_validation      = true
  }
}
```

**Note**: You can use Spaces itself as a Terraform backend!

### Multi-Environment Pattern

Use workspaces or separate directories:

```bash
# Using workspaces
terraform workspace new production
terraform workspace select production
terraform apply

# Or separate directories
./environments/
  dev/main.tf
  staging/main.tf
  production/main.tf
```

## Troubleshooting

### Issue: "Bucket name already in use"
**Solution**: Bucket names must be globally unique across all DigitalOcean Spaces. Choose a different name.

### Issue: "Region validation failed"
**Solution**: Ensure region is a valid DigitalOcean Spaces region (nyc3, sfo3, ams3, sgp1, etc.)

### Issue: "Access denied"
**Solution**: 
- Verify DigitalOcean API token has Spaces permissions
- For object access, configure `spaces_access_id` and `spaces_secret_key`

### Issue: "Versioning cannot be disabled"
**Note**: This is by design. Once enabled, versioning can only be suspended, not fully disabled.

### Issue: "Invalid bucket name"
**Solution**: Bucket names must be DNS-compatible:
- 3-63 characters
- Lowercase alphanumeric and hyphens only
- Cannot start or end with hyphen

## Accessing Created Buckets

### Using AWS CLI

```bash
# Configure AWS CLI for Spaces
aws configure set aws_access_key_id $SPACES_KEY
aws configure set aws_secret_access_key $SPACES_SECRET

# List objects
aws --endpoint-url https://nyc3.digitaloceanspaces.com s3 ls s3://my-bucket/

# Upload
aws --endpoint-url https://nyc3.digitaloceanspaces.com s3 cp file.txt s3://my-bucket/

# Download
aws --endpoint-url https://nyc3.digitaloceanspaces.com s3 cp s3://my-bucket/file.txt ./
```

### Using Terraform Outputs in Application

```hcl
# Export bucket endpoint to app configuration
resource "kubernetes_config_map" "app_config" {
  metadata {
    name = "app-config"
  }

  data = {
    BUCKET_ENDPOINT = module.app_bucket.endpoint
    BUCKET_NAME     = module.app_bucket.bucket_name
    BUCKET_REGION   = module.app_bucket.region
  }
}
```

## Examples

For comprehensive examples including multi-region setups, versioning, and production configurations, see:
- [API-level examples](../../examples.md)
- [Terraform-specific examples](./examples.md) (if available)

## References

- [Terraform DigitalOcean Provider](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs)
- [Spaces Bucket Resource](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/spaces_bucket)
- [DigitalOcean Spaces Documentation](https://docs.digitalocean.com/products/spaces/)

## Support

- **Issues**: [github.com/plantonhq/planton/issues](https://github.com/plantonhq/planton/issues)
- **DigitalOcean Support**: [digitalocean.com/support](https://www.digitalocean.com/support/)

