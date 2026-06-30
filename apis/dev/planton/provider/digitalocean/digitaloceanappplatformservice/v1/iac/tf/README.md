# DigitalOcean App Platform Service - Terraform Module

## Overview

This Terraform module deploys containerized applications on DigitalOcean App Platform using a declarative configuration. The module supports web services, workers, and jobs with git-based or container image sources, autoscaling, custom domains, and environment management.

## Features

- **Multiple Service Types**: Web services (HTTP traffic), workers (background processing), and jobs (pre-deployment tasks)
- **Source Flexibility**: Deploy from Git repositories or container images
- **Autoscaling**: CPU-based horizontal autoscaling for web services
- **Custom Domains**: Automatic SSL certificate provisioning
- **Environment Variables**: Secure runtime configuration
- **Type-Safe**: Terraform variable validation enforces correct configurations

## Prerequisites

- **Terraform**: Version 1.0 or higher
- **DigitalOcean Account**: With API access
- **DigitalOcean API Token**: For provider authentication

## Installation

This module is part of the Planton monorepo:

```bash
git clone https://github.com/plantonhq/planton.git
cd planton/apis/dev/planton/provider/digitalocean/digitaloceanappplatformservice/v1/iac/tf
```

## Usage

### Basic Example

```hcl
module "web_service" {
  source = "./path/to/module"

  metadata = {
    name = "my-api"
    labels = {
      environment = "production"
      team        = "backend"
    }
  }

  spec = {
    service_name = "my-api"
    region       = "nyc1"
    service_type = "web_service"
    
    git_source = {
      repo_url = "https://github.com/myorg/my-api.git"
      branch   = "main"
    }
    
    instance_size_slug = "basic-xxs"
    instance_count     = 1
  }

  digitalocean_token = var.digitalocean_token
}

output "live_url" {
  value = module.web_service.live_url
}
```

### Production Example with Autoscaling

```hcl
module "production_api" {
  source = "./path/to/module"

  metadata = {
    name = "production-api"
    labels = {
      environment = "production"
      criticality = "high"
    }
  }

  spec = {
    service_name = "production-api"
    region       = "sfo3"
    service_type = "web_service"
    
    git_source = {
      repo_url      = "https://github.com/myorg/api.git"
      branch        = "production"
      build_command = "npm run build"
      run_command   = "npm start"
    }
    
    instance_size_slug = "professional-s"
    enable_autoscale   = true
    min_instance_count = 2
    max_instance_count = 10
    
    env = {
      NODE_ENV   = "production"
      LOG_LEVEL  = "info"
    }
    
    custom_domain = "api.myapp.com"
  }

  digitalocean_token = var.digitalocean_token
}
```

## Module Inputs

### Required Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Resource metadata (name, labels) |
| `spec` | object | Service specification (see below) |
| `digitalocean_token` | string | DigitalOcean API token (sensitive) |

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `service_name` | string | Yes | Unique service name (DNS-friendly) |
| `region` | string | Yes | DigitalOcean region (e.g., `nyc1`) |
| `service_type` | string | Yes | `web_service`, `worker`, or `job` |
| `git_source` | object | Conditional | Git repository source |
| `image_source` | object | Conditional | Container image source |
| `instance_size_slug` | string | Yes | Instance size (e.g., `basic-xxs`) |
| `instance_count` | number | No | Number of instances (default: 1) |
| `enable_autoscale` | bool | No | Enable autoscaling (default: false) |
| `min_instance_count` | number | Conditional | Min instances for autoscaling |
| `max_instance_count` | number | Conditional | Max instances for autoscaling |
| `env` | map(string) | No | Environment variables |
| `custom_domain` | string | No | Custom domain name |

## Module Outputs

| Output | Description |
|--------|-------------|
| `app_id` | DigitalOcean App ID |
| `live_url` | Live URL of the application |
| `default_ingress` | Default ingress URL |
| `app_urn` | Uniform Resource Name of the app |
| `active_deployment_id` | Current deployment ID |
| `created_at` | App creation timestamp |
| `updated_at` | Last update timestamp |

## Validation Rules

The module includes automatic validation:

1. **Service Type**: Must be `web_service`, `worker`, or `job`
2. **Source Configuration**: Exactly one of `git_source` or `image_source` required
3. **Autoscaling**: If enabled, both `min_instance_count` and `max_instance_count` must be set
4. **Autoscaling Type**: Only supported for `web_service` type

## Service Types

### Web Service
- Receives HTTP traffic
- Supports autoscaling
- Gets a public URL
- Load balanced automatically

### Worker
- Background processing only
- No HTTP ingress
- Fixed instance count
- Ideal for queue workers

### Job
- Runs before deployments
- Pre-deployment tasks only
- Single instance
- Ideal for migrations

## Source Configuration

### Git Source

Deploy from a Git repository:

```hcl
git_source = {
  repo_url      = "https://github.com/myorg/myapp.git"
  branch        = "main"
  build_command = "npm run build"  # Optional
  run_command   = "npm start"      # Optional
}
```

### Image Source

Deploy from a container registry:

```hcl
image_source = {
  registry   = "registry.digitalocean.com/myorg"
  repository = "myapp"
  tag        = "v1.0.0"
}
```

## Instance Sizes

| Size | Monthly Cost | RAM | CPU | Use Case |
|------|--------------|-----|-----|----------|
| `basic-xxs` | $5 | 512MB | Shared | Dev/Test |
| `basic-xs` | $12 | 1GB | Shared | Small apps |
| `basic-s` | $24 | 2GB | Shared | Moderate traffic |
| `professional-xs` | $12 | 1GB | 1 vCPU | Production (small) |
| `professional-s` | $24 | 2GB | 1 vCPU | Production (medium) |
| `professional-m` | $48 | 4GB | 2 vCPU | Production (large) |

## Terraform Workflow

### Initialize

```bash
terraform init
```

### Plan

```bash
terraform plan -var="digitalocean_token=$DIGITALOCEAN_TOKEN"
```

### Apply

```bash
terraform apply -var="digitalocean_token=$DIGITALOCEAN_TOKEN"
```

### Destroy

```bash
terraform destroy -var="digitalocean_token=$DIGITALOCEAN_TOKEN"
```

## Best Practices

### Credential Management

**Never commit tokens to version control**. Use one of these methods:

#### Environment Variables

```bash
export TF_VAR_digitalocean_token="your-token"
terraform apply
```

#### Terraform Cloud

Store `digitalocean_token` as a sensitive workspace variable.

#### HashiCorp Vault

```hcl
data "vault_generic_secret" "do_token" {
  path = "secret/digitalocean/token"
}

module "app" {
  source = "./path/to/module"
  digitalocean_token = data.vault_generic_secret.do_token.data["token"]
  # ... other config ...
}
```

### State Management

Use remote state for production:

```hcl
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "digitalocean-apps/prod/terraform.tfstate"
    region = "us-east-1"
  }
}
```

### Multi-Environment Pattern

Use workspaces or separate directories:

```bash
# Using workspaces
terraform workspace new production
terraform workspace select production
terraform apply

# Or separate directories
./environments/
  dev/
    main.tf
  staging/
    main.tf
  production/
    main.tf
```

## Troubleshooting

### Issue: "Invalid service_type"
**Solution**: Ensure `service_type` is exactly one of: `web_service`, `worker`, `job`

### Issue: "Must specify exactly one of git_source or image_source"
**Solution**: Provide either `git_source` OR `image_source`, not both

### Issue: "When enable_autoscale is true, both min_instance_count and max_instance_count must be specified"
**Solution**: Set both min and max instance counts when autoscaling is enabled

### Issue: "Autoscaling is only supported for web_service type"
**Solution**: Remove `enable_autoscale` for worker or job service types

## Examples

For comprehensive examples including autoscaling, custom domains, and multi-environment setups, see:
- [Terraform examples.md](./examples.md)
- [API-level examples](../../examples.md)

## References

- [Terraform DigitalOcean Provider](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs)
- [DigitalOcean App Spec Reference](https://docs.digitalocean.com/products/app-platform/reference/app-spec/)
- [Planton Documentation](https://github.com/plantonhq/planton)

## Support

- **Issues**: [github.com/plantonhq/planton/issues](https://github.com/plantonhq/planton/issues)
- **DigitalOcean Support**: [digitalocean.com/support](https://www.digitalocean.com/support/)

