# DigitalOcean Load Balancer - Terraform Module

This Terraform module deploys DigitalOcean Regional Load Balancers from Planton's protobuf-defined manifests.

## Overview

The module translates `DigitalOceanLoadBalancerSpec` manifests into DigitalOcean Load Balancer resources using Terraform's DigitalOcean provider. It handles:

- Regional load balancer provisioning  
- VPC network placement
- Tag-based or ID-based Droplet targeting
- Forwarding rules with SSL termination support
- Health check configuration
- Sticky sessions (optional)

## Prerequisites

### 1. Terraform

Install Terraform 1.5 or later:

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

### 2. DigitalOcean API Token

Get your token from: https://cloud.digitalocean.com/account/api/tokens

```bash
export DIGITALOCEAN_TOKEN="your-api-token"
```

### 3. VPC and Droplets

- A DigitalOcean VPC must exist in the target region
- Droplets must be created and tagged (for tag-based targeting) or have known IDs

## Module Structure

```
iac/tf/
├── variables.tf    # Input variable definitions
├── provider.tf     # Terraform and provider configuration
├── locals.tf       # Local variables and computed values
├── main.tf         # Load balancer resource
├── outputs.tf      # Module outputs
└── README.md       # This file
```

## Input Variables

### Required Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Resource metadata (name, org, env, labels, tags) |
| `spec` | object | Load balancer specification (see below) |

### Spec Object Structure

```hcl
spec = {
  load_balancer_name = string              # Name (lowercase alphanumeric + hyphens)
  region             = string              # DigitalOcean region (e.g., "nyc3", "sfo3")
  vpc                = object({            # VPC reference
    value = string                         # VPC UUID
    ref   = string                         # Reference to VPC resource (future)
  })
  forwarding_rules = list(object({         # Traffic routing rules
    entry_port       = number              # Load balancer port (1-65535)
    entry_protocol   = string              # "http", "https", or "tcp"
    target_port      = number              # Backend port (1-65535)
    target_protocol  = string              # "http", "https", or "tcp"
    certificate_name = string              # SSL certificate name (for HTTPS)
  }))
  health_check = object({                  # Health check configuration
    port              = number             # Health check port
    protocol          = string             # "http", "https", or "tcp"
    path              = string             # Health check path (HTTP/HTTPS only)
    check_interval_sec = number            # Check interval in seconds (default: 10)
  })
  droplet_ids = list(object({              # Specific Droplet IDs (optional)
    value = string
    ref   = string
  }))
  droplet_tag            = string          # Droplet tag for dynamic targeting (optional)
  enable_sticky_sessions = bool            # Enable cookie-based session affinity
}
```

## Usage

### Example: Basic HTTP Load Balancer

Create a `terraform.tfvars` file:

```hcl
metadata = {
  name = "dev-web-lb"
  org  = "my-org"
  env  = "development"
  labels = {
    app = "web"
  }
}

spec = {
  load_balancer_name = "dev-web-lb"
  region             = "nyc3"
  
  vpc = {
    value = "vpc-123456"
  }
  
  droplet_tag = "web-dev"
  
  forwarding_rules = [
    {
      entry_port      = 80
      entry_protocol  = "http"
      target_port     = 80
      target_protocol = "http"
      certificate_name = null
    }
  ]
  
  health_check = {
    port              = 80
    protocol          = "tcp"
    path              = null
    check_interval_sec = 10
  }
  
  enable_sticky_sessions = false
}
```

Deploy:

```bash
terraform init
terraform plan
terraform apply
```

### Example: Production HTTPS Load Balancer

```hcl
spec = {
  load_balancer_name = "prod-web-lb"
  region             = "sfo3"
  
  vpc = {
    value = "vpc-prod-789"
  }
  
  droplet_tag = "web-prod"
  
  forwarding_rules = [
    {
      entry_port       = 443
      entry_protocol   = "https"
      target_port      = 80
      target_protocol  = "http"
      certificate_name = "my-le-cert-name"  # Use certificate NAME, not ID
    }
  ]
  
  health_check = {
    port              = 80
    protocol          = "http"
    path              = "/healthz"
    check_interval_sec = 10
  }
  
  enable_sticky_sessions = true
}
```

### Example: Multi-Port (HTTP + HTTPS)

```hcl
spec = {
  load_balancer_name = "dual-port-lb"
  region             = "fra1"
  
  vpc = {
    value = "vpc-eu-456"
  }
  
  droplet_tag = "web-dual"
  
  forwarding_rules = [
    # HTTP rule
    {
      entry_port      = 80
      entry_protocol  = "http"
      target_port     = 8080
      target_protocol = "http"
      certificate_name = null
    },
    # HTTPS rule
    {
      entry_port       = 443
      entry_protocol   = "https"
      target_port      = 8080
      target_protocol  = "http"
      certificate_name = "dual-cert"
    }
  ]
  
  health_check = {
    port              = 8080
    protocol          = "http"
    path              = "/health"
    check_interval_sec = 10
  }
}
```

### Example: TCP Load Balancer for Database

```hcl
spec = {
  load_balancer_name = "prod-mysql-lb"
  region             = "nyc1"
  
  vpc = {
    value = "vpc-db-789"
  }
  
  droplet_tag = "mysql-galera"
  
  forwarding_rules = [
    {
      entry_port      = 3306
      entry_protocol  = "tcp"
      target_port     = 3306
      target_protocol = "tcp"
      certificate_name = null
    }
  ]
  
  health_check = {
    port              = 3306
    protocol          = "tcp"
    path              = null
    check_interval_sec = 10
  }
}
```

## Outputs

| Output | Description |
|--------|-------------|
| `load_balancer_id` | Load balancer UUID |
| `load_balancer_ip` | Public IP address |
| `load_balancer_urn` | Uniform resource name |
| `load_balancer_name` | Load balancer name |
| `status` | Load balancer status (new, active, errored) |
| `outputs` | Complete outputs object for cross-stack references |

Access outputs:

```bash
terraform output load_balancer_ip
terraform output -json outputs
```

Reference in other modules:

```hcl
data "terraform_remote_state" "lb" {
  backend = "s3"
  config = {
    bucket = "my-terraform-state"
    key    = "load-balancer/terraform.tfstate"
    region = "us-east-1"
  }
}

# Use load balancer IP in DNS record
resource "digitalocean_record" "www" {
  domain = "example.com"
  type   = "A"
  name   = "www"
  value  = data.terraform_remote_state.lb.outputs.outputs.ip
}
```

## State Management

### Local State (Development)

For testing and development, use local state:

```bash
terraform init
terraform apply
```

State is stored in `terraform.tfstate`.

### Remote State (Production)

For production, use remote state with locking:

#### Option 1: DigitalOcean Spaces (S3-Compatible)

```hcl
# backend.tf
terraform {
  backend "s3" {
    endpoint                    = "nyc3.digitaloceanspaces.com"
    region                      = "us-east-1"  # Dummy value (required by provider)
    bucket                      = "my-terraform-state"
    key                         = "digitalocean-lb/terraform.tfstate"
    skip_credentials_validation = true
    skip_metadata_api_check     = true
  }
}
```

Configure credentials:

```bash
export AWS_ACCESS_KEY_ID="your-spaces-key"
export AWS_SECRET_ACCESS_KEY="your-spaces-secret"
```

Initialize:

```bash
terraform init
```

#### Option 2: Terraform Cloud

```hcl
# backend.tf
terraform {
  backend "remote" {
    organization = "my-org"
    workspaces {
      name = "digitalocean-lb-prod"
    }
  }
}
```

## Workflow

### Initial Deployment

```bash
# Initialize Terraform
terraform init

# Validate configuration
terraform validate

# Preview changes
terraform plan

# Apply changes
terraform apply

# Save outputs
terraform output -json > outputs.json
```

### Updates

```bash
# Modify terraform.tfvars or spec

# Preview changes
terraform plan

# Apply updates
terraform apply
```

### Destruction

```bash
# Preview destruction
terraform plan -destroy

# Destroy infrastructure
terraform destroy
```

## Best Practices

### 1. Use Tag-Based Targeting

✅ **Recommended:**
```hcl
spec = {
  droplet_tag = "web-prod"
}
```

**Benefits:**
- Automatic backend discovery
- Works with autoscaling
- Enables blue-green deployments

❌ **Avoid (unless necessary):**
```hcl
spec = {
  droplet_ids = [
    { value = "386734086" },
    { value = "386734087" }
  ]
}
```

**Drawbacks:**
- Manual management
- Doesn't scale
- Requires updates for every backend change

### 2. Use Certificate Names, Not IDs

✅ **Correct:**
```hcl
certificate_name = "my-le-cert-name"
```

❌ **Wrong:**
```hcl
certificate_name = "a1b2c3d4-e5f6-7890-abcd-1234567890ab"  # ID changes on renewal
```

**Reason:** Let's Encrypt certificate IDs change on auto-renewal, breaking Terraform state.

### 3. Implement Proper Health Checks

✅ **Application-level health check:**
```hcl
health_check = {
  port     = 80
  protocol = "http"
  path     = "/healthz"
}
```

**Backend implementation:**
```go
http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    // Check database connectivity
    if db.Ping() != nil {
        w.WriteHeader(500)
        return
    }
    w.WriteHeader(200)
})
```

❌ **Avoid TCP-only checks (if HTTP is available):**
```hcl
health_check = {
  protocol = "tcp"  # Only checks if port is open, not if app is healthy
}
```

### 4. Use VPC for Private Communication

```hcl
spec = {
  vpc = {
    value = "vpc-123456"  # Place LB and Droplets in same VPC
  }
}
```

**Benefits:**
- Unmetered traffic between LB and Droplets
- Secure private network communication
- Use Cloud Firewalls to block direct Droplet access

### 5. Lifecycle Management

The module includes best-practice lifecycle rules:

```hcl
lifecycle {
  prevent_destroy = false  # Set to true for production
  
  # Ignore Droplet ID changes when using tag-based targeting
  ignore_changes = [
    droplet_ids
  ]
}
```

## Troubleshooting

### "No healthy backends" (503 Error)

**Symptoms:** Load balancer returns 503 Service Unavailable

**Causes:**
1. Health check misconfiguration (wrong port, path, or protocol)
2. Backend application not responding to health checks
3. Droplets not tagged correctly
4. Droplets in different VPC or region

**Solutions:**

```bash
# Check Droplet tags
doctl compute droplet list --format ID,Name,Tags

# Test health check endpoint on Droplet
ssh droplet-ip
curl http://localhost:80/healthz

# Check load balancer status
doctl compute load-balancer get <load_balancer_id>

# Verify forwarding rules
terraform show
```

### "Certificate not found"

**Symptoms:** Terraform fails with "certificate does not exist"

**Cause:** Certificate name is incorrect or doesn't exist

**Solution:**

```bash
# List certificates
doctl compute certificate list

# Use the NAME column value (not ID)
```

### State Drift with Droplet IDs

**Symptoms:** Terraform constantly wants to update `droplet_ids`

**Cause:** Tag-based targeting dynamically updates backend pool

**Solution:** This is expected behavior. The `ignore_changes` lifecycle rule prevents unnecessary updates.

### Provider Authentication Errors

**Symptoms:** "Invalid token" or "Unauthorized"

**Solution:**

```bash
# Verify token is set
echo $DIGITALOCEAN_TOKEN

# Test token with doctl
doctl auth init --access-token $DIGITALOCEAN_TOKEN
doctl account get
```

## Advanced Configuration

### Multiple Environments with Workspaces

```bash
# Create dev workspace
terraform workspace new dev

# Create prod workspace
terraform workspace new prod

# Switch to prod
terraform workspace select prod

# Apply with environment-specific variables
terraform apply -var-file="prod.tfvars"
```

### Blue-Green Deployment

```hcl
# lb-blue.tfvars
spec = {
  droplet_tag = "blue"
  # ... other config
}

# Deploy green environment
# Update tfvars to switch tag to "green"
spec = {
  droplet_tag = "green"
}

# Apply - instant cutover
terraform apply
```

### Custom Backend Configuration

```bash
# Initialize with backend config file
terraform init -backend-config=backend-prod.hcl

# backend-prod.hcl
endpoint = "nyc3.digitaloceanspaces.com"
bucket   = "my-terraform-state"
key      = "prod/digitalocean-lb/terraform.tfstate"
```

## Integration with Planton

This Terraform module can be used standalone or as part of Planton's manifest-driven workflow:

```bash
# Convert manifest to Terraform variables
planton convert --manifest lb.yaml --output tfvars

# Deploy with Terraform
terraform apply -var-file=lb.tfvars
```

**Note:** Planton primarily uses Pulumi for deployments. This Terraform module provides an alternative for teams preferring Terraform.

## Validation

### Pre-Deployment

```bash
# Format code
terraform fmt -recursive

# Validate configuration
terraform validate

# Security scan
tfsec .

# Preview changes
terraform plan
```

### Post-Deployment

```bash
# Get load balancer IP
LB_IP=$(terraform output -raw load_balancer_ip)

# Test HTTP endpoint
curl http://$LB_IP

# Test HTTPS endpoint
curl https://$LB_IP

# Check health
curl http://$LB_IP/healthz
```

## Production Checklist

Before deploying to production:

- [ ] VPC created in target region
- [ ] Droplets tagged correctly
- [ ] Certificate uploaded (for HTTPS)
- [ ] Health check endpoint implemented on backends
- [ ] Cloud Firewalls configured (block direct Droplet access)
- [ ] Remote state backend configured
- [ ] State locking enabled (Terraform Cloud or DynamoDB)
- [ ] Monitoring and alerting configured
- [ ] DNS A record ready to update
- [ ] Tested in staging environment

## Cost Estimation

DigitalOcean Load Balancer pricing:

- **Layer 7 (HTTP/HTTPS):** $12/month per node
- **Layer 4 (TCP):** $15/month per node

**Default:** 1 node (sufficient for most use cases)

**Estimate cost:**

```bash
# HTTP/HTTPS load balancer
terraform plan | grep "digitalocean_loadbalancer"
# Cost: $12/month

# TCP load balancer
# Cost: $15/month
```

Use `terraform-cost-estimation` or Infracost for detailed cost analysis.

## Next Steps

- Review [../../docs/README.md](../../docs/README.md) for architecture and best practices
- Check [../../examples.md](../../examples.md) for usage patterns
- See [../pulumi/README.md](../pulumi/README.md) for Pulumi alternative
- See [../../hack/manifest.yaml](../../hack/manifest.yaml) for test manifest

## Support

For issues or questions:
- Check [troubleshooting section](#troubleshooting)
- Review [Terraform DigitalOcean Provider docs](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/loadbalancer)
- Review [DigitalOcean Load Balancer docs](https://docs.digitalocean.com/products/networking/load-balancers/)
- Open an issue in the Planton repository

