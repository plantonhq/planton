# Terraform Module: Cloudflare Load Balancer

This directory contains the Terraform module for deploying Cloudflare Load Balancers.

## Overview

The Terraform module provisions three Cloudflare resources:

1. **cloudflare_load_balancer_monitor** (account-level): Health check configuration
2. **cloudflare_load_balancer_pool** (account-level): Group of origin servers
3. **cloudflare_load_balancer** (zone-level): DNS hostname and routing policy

## Module Structure

```
iac/tf/
├── README.md        # This file - deployment guide
├── variables.tf     # Input variables
├── provider.tf      # Cloudflare provider configuration
├── locals.tf        # Local variables and computed values
├── main.tf          # Resource definitions
└── outputs.tf       # Output values
```

## Prerequisites

1. **Cloudflare Account**:
   - Active Cloudflare account with Load Balancing add-on enabled
   - Cloudflare API token with permissions:
     - Load Balancers: Edit
     - Zones: Read
     - Account Settings: Read

2. **Required Environment Variables**:
   ```bash
   export CLOUDFLARE_API_TOKEN="your-api-token-here"
   ```

3. **Terraform CLI**:
   ```bash
   # macOS
   brew install terraform

   # Linux
   wget https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip
   unzip terraform_1.6.0_linux_amd64.zip
   sudo mv terraform /usr/local/bin/

   # Verify installation
   terraform version
   ```

## Usage

### Step 1: Create Terraform Configuration

Create a `main.tf` file using this module:

```hcl
terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

module "cloudflare_load_balancer" {
  source = "./iac/tf"

  metadata = {
    name = "api-lb"
    labels = {
      env     = "production"
      service = "api"
    }
  }

  spec = {
    hostname = "api.example.com"
    
    zone_id = {
      value = "abc123def456"  # Your Cloudflare zone ID
    }
    
    origins = [
      {
        name    = "primary"
        address = "203.0.113.10"
        weight  = 1
      },
      {
        name    = "secondary"
        address = "198.51.100.20"
        weight  = 1
      }
    ]
    
    proxied           = true
    health_probe_path = "/health"
    session_affinity  = 1  # SESSION_AFFINITY_COOKIE
    steering_policy   = 0  # STEERING_OFF (active-passive failover)
  }
}

output "load_balancer_id" {
  value = module.cloudflare_load_balancer.load_balancer_id
}
```

### Step 2: Initialize Terraform

```bash
cd iac/tf

# Initialize providers and modules
terraform init
```

### Step 3: Plan Deployment

```bash
# Preview changes
terraform plan
```

**Expected Output**:

```
Terraform will perform the following actions:

  # cloudflare_load_balancer.main will be created
  + resource "cloudflare_load_balancer" "main" {
      + id               = (known after apply)
      + name             = "api.example.com"
      + zone_id          = "abc123def456"
      + proxied          = true
      + steering_policy  = "off"
      + session_affinity = "cookie"
    }

  # cloudflare_load_balancer_pool.main will be created
  + resource "cloudflare_load_balancer_pool" "main" {
      + id      = (known after apply)
      + name    = "api-lb-pool"
      + enabled = true
    }

  # cloudflare_load_balancer_monitor.health_check will be created
  + resource "cloudflare_load_balancer_monitor" "health_check" {
      + id            = (known after apply)
      + type          = "https"
      + path          = "/health"
      + expected_codes = "2xx"
    }

Plan: 3 to add, 0 to change, 0 to destroy.
```

### Step 4: Apply Configuration

```bash
# Apply changes
terraform apply

# Auto-approve (for CI/CD)
terraform apply -auto-approve
```

### Step 5: Verify Deployment

```bash
# View outputs
terraform output

# Test the load balancer
curl https://api.example.com/health
```

## Input Variables

### Required Variables

#### `metadata` (object)

Metadata for the load balancer resource.

```hcl
metadata = {
  name = "api-lb"  # Required
}
```

#### `spec` (object)

Load balancer specification.

**Required fields**:

- **`hostname`** (string): DNS hostname for the load balancer
  ```hcl
  hostname = "api.example.com"
  ```

- **`zone_id`** (object): Cloudflare zone ID
  ```hcl
  zone_id = {
    value = "abc123def456"
  }
  ```

- **`origins`** (list of objects): Origin servers (minimum 1 required)
  ```hcl
  origins = [
    {
      name    = "primary"
      address = "203.0.113.10"
      weight  = 1  # Optional, defaults to 1
    }
  ]
  ```

**Optional fields**:

- **`proxied`** (bool): Enable Cloudflare proxy - Default: `true`
  ```hcl
  proxied = true  # Orange cloud mode
  ```

- **`health_probe_path`** (string): Health check path - Default: `"/"`
  ```hcl
  health_probe_path = "/healthz"
  ```

- **`session_affinity`** (number): Session stickiness - Default: `0` (none)
  - `0` = SESSION_AFFINITY_NONE
  - `1` = SESSION_AFFINITY_COOKIE
  ```hcl
  session_affinity = 1
  ```

- **`steering_policy`** (number): Traffic routing policy - Default: `0` (failover)
  - `0` = STEERING_OFF (active-passive failover)
  - `1` = STEERING_GEO (geographic routing)
  - `2` = STEERING_RANDOM (weighted distribution)
  ```hcl
  steering_policy = 0
  ```

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `load_balancer_id` | string | Cloudflare Load Balancer resource ID |
| `load_balancer_dns_record_name` | string | DNS hostname (e.g., `api.example.com`) |
| `load_balancer_cname_target` | string | CNAME target for external DNS |
| `pool_id` | string | Load balancer pool resource ID |
| `monitor_id` | string | Health check monitor resource ID |

Access outputs:

```bash
# View all outputs
terraform output

# Get specific output
terraform output load_balancer_id
```

Use outputs in other modules:

```hcl
module "other_module" {
  source = "./other-module"
  
  load_balancer_id = module.cloudflare_load_balancer.load_balancer_id
}
```

## Examples

### Example 1: Basic Active-Passive Failover

```hcl
module "lb_failover" {
  source = "./iac/tf"

  metadata = {
    name = "api-failover-lb"
  }

  spec = {
    hostname = "api.example.com"
    zone_id  = { value = "abc123" }
    
    steering_policy = 0  # STEERING_OFF (failover)
    
    origins = [
      { name = "primary",   address = "203.0.113.10", weight = 1 },
      { name = "secondary", address = "198.51.100.20", weight = 1 }
    ]
    
    proxied           = true
    health_probe_path = "/"
    session_affinity  = 0  # No sticky sessions
  }
}
```

### Example 2: Geographic Routing

```hcl
module "lb_geo" {
  source = "./iac/tf"

  metadata = {
    name = "global-app-lb"
  }

  spec = {
    hostname = "app.example.com"
    zone_id  = { value = "abc123" }
    
    steering_policy = 1  # STEERING_GEO
    
    origins = [
      { name = "us-east",   address = "203.0.113.10", weight = 1 },
      { name = "eu-west",   address = "198.51.100.20", weight = 1 },
      { name = "ap-south",  address = "192.0.2.30", weight = 1 }
    ]
    
    proxied           = true
    health_probe_path = "/health"
    session_affinity  = 1  # Cookie-based sticky sessions
  }
}
```

### Example 3: Weighted A/B Testing

```hcl
module "lb_canary" {
  source = "./iac/tf"

  metadata = {
    name = "canary-deployment-lb"
  }

  spec = {
    hostname = "app.example.com"
    zone_id  = { value = "abc123" }
    
    steering_policy = 2  # STEERING_RANDOM (weighted)
    
    origins = [
      { name = "stable", address = "203.0.113.10", weight = 90 },  # 90% traffic
      { name = "canary", address = "198.51.100.20", weight = 10 }  # 10% traffic
    ]
    
    proxied           = true
    health_probe_path = "/healthz"
    session_affinity  = 1  # Ensure users stick to same version
  }
}
```

## State Management

### Local State (Development Only)

By default, Terraform stores state locally in `terraform.tfstate`:

```bash
terraform apply  # Creates terraform.tfstate
```

**Warning**: Do not use local state for production. State files contain sensitive data.

### Remote State (Recommended)

#### S3 Backend

```hcl
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "cloudflare/load-balancer/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-lock"
  }
}
```

#### Terraform Cloud

```hcl
terraform {
  cloud {
    organization = "my-org"
    workspaces {
      name = "cloudflare-load-balancer"
    }
  }
}
```

## Updating the Load Balancer

Modify your Terraform configuration and re-run:

```bash
terraform plan   # Preview changes
terraform apply  # Apply changes
```

### Common Updates

**Add a new origin**:

```hcl
origins = [
  { name = "primary",   address = "203.0.113.10" },
  { name = "secondary", address = "198.51.100.20" },
  { name = "tertiary",  address = "192.0.2.30" }  # Added
]
```

**Change health check path**:

```hcl
health_probe_path = "/api/v1/healthz"
```

**Enable session affinity**:

```hcl
session_affinity = 1  # SESSION_AFFINITY_COOKIE
```

Terraform will show a diff before applying:

```
  ~ resource "cloudflare_load_balancer" "main" {
      ~ session_affinity = "none" -> "cookie"
    }
```

## Destroying the Load Balancer

```bash
# Preview what will be deleted
terraform plan -destroy

# Confirm and delete all resources
terraform destroy
```

**Warning**: This permanently deletes the load balancer. Ensure no production traffic depends on it.

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy Cloudflare Load Balancer

on:
  push:
    branches: [main]

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: hashicorp/setup-terraform@v2
        with:
          terraform_version: 1.6.0
      
      - name: Terraform Init
        run: terraform init
        working-directory: iac/tf
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
      
      - name: Terraform Plan
        run: terraform plan
        working-directory: iac/tf
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
      
      - name: Terraform Apply
        run: terraform apply -auto-approve
        working-directory: iac/tf
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
```

## Troubleshooting

### Common Issues

**Issue**: `Error: zone_id is required`

**Solution**: Ensure `zone_id` is set in the spec:
```hcl
zone_id = { value = "your-zone-id" }
```

---

**Issue**: `Error: authentication error - invalid API token`

**Solution**: Verify `CLOUDFLARE_API_TOKEN` environment variable:
```bash
echo $CLOUDFLARE_API_TOKEN
```

---

**Issue**: Origins show as "Unhealthy" in Cloudflare dashboard

**Solution**:
1. Check origin servers are running
2. Verify `health_probe_path` returns HTTP 200
3. Ensure origin firewalls allow Cloudflare health check IPs

---

**Issue**: `Error: resource already exists`

**Solution**: Import existing resource:
```bash
terraform import cloudflare_load_balancer.main <load-balancer-id>
```

### Debug Mode

Enable detailed logging:

```bash
export TF_LOG=DEBUG
terraform apply
```

### View Resource State

```bash
# List all resources in state
terraform state list

# Show details of a specific resource
terraform state show cloudflare_load_balancer.main
```

## Best Practices

1. **Use remote state backend** (S3, Terraform Cloud) for production
2. **Store API tokens in secrets** (GitHub Secrets, AWS Secrets Manager)
3. **Separate environments** (dev, staging, prod) using workspaces or separate state files
4. **Enable state locking** (DynamoDB for S3 backend) to prevent concurrent modifications
5. **Review plans carefully** before applying to production
6. **Tag resources** using `metadata.labels` for cost tracking and organization

## Terraform Workspaces

Manage multiple environments with workspaces:

```bash
# Create dev workspace
terraform workspace new dev
terraform apply

# Create prod workspace
terraform workspace new prod
terraform apply

# Switch between workspaces
terraform workspace select dev
terraform workspace select prod
```

Each workspace maintains separate state.

## Validation

Validate configuration before applying:

```bash
# Check syntax
terraform fmt -check

# Validate configuration
terraform validate

# Security scan (requires tfsec)
tfsec .
```

## Module Versioning

When using this module from a remote source, pin to a specific version:

```hcl
module "cloudflare_load_balancer" {
  source = "git::https://github.com/org/repo.git//iac/tf?ref=v1.0.0"
  
  # ...
}
```

## Additional Resources

- [Cloudflare Terraform Provider Docs](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs)
- [Cloudflare Load Balancer API Docs](https://developers.cloudflare.com/api/operations/load-balancers-create-load-balancer)
- [Component README](../../README.md) - User-facing documentation
- [Examples](../../examples.md) - Complete usage examples

## Support

For issues or questions:
1. Check [Common Issues](#common-issues) above
2. Review [Component README](../../README.md)
3. Consult Cloudflare and Terraform official documentation

---

**Ready to deploy?** Run `terraform init && terraform apply` to get started!

