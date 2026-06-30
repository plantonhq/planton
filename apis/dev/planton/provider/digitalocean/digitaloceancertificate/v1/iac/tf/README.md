# DigitalOcean Certificate - Terraform Module

This Terraform module provisions and manages SSL/TLS certificates on DigitalOcean using the official `digitalocean` provider.

## Overview

The module implements Planton's `DigitalOceanCertificate` protobuf spec, providing a declarative interface for:

- **Let's Encrypt Certificates**: Free, auto-renewing certificates (requires DigitalOcean DNS)
- **Custom Certificates**: User-provided certificates (BYOC - Bring Your Own Certificate)

## Module Structure

```
iac/tf/
├── variables.tf    # Input variable definitions
├── locals.tf       # Local value transformations
├── main.tf         # Certificate resource definition
├── outputs.tf      # Output value exports
├── provider.tf     # Provider configuration
└── README.md       # This file
```

## Prerequisites

- **Terraform**: Version 1.0 or later
- **DigitalOcean API Token**: Set via environment variable `DIGITALOCEAN_TOKEN` or provider configuration
- **DigitalOcean Provider**: Version 2.x (automatically downloaded by Terraform)

## Quick Start

### 1. Set Up Environment

```bash
# Navigate to the Terraform module directory
cd apis/dev/planton/provider/digitalocean/digitaloceancertificate/v1/iac/tf

# Set your DigitalOcean API token
export DIGITALOCEAN_TOKEN="your-digitalocean-api-token"
```

### 2. Create Terraform Configuration

**Let's Encrypt Example:**
```hcl
module "prod_web_cert" {
  source = "./path/to/digitaloceancertificate/v1/iac/tf"

  metadata = {
    name = "prod-web-cert"
    id   = "docert-abc123"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    certificate_name = "prod-web-cert"
    type             = "letsEncrypt"
    lets_encrypt = {
      domains = [
        "example.com",
        "www.example.com"
      ]
    }
    description = "Production web certificate"
    tags        = ["env:production", "team:platform"]
  }
}
```

**Custom Certificate Example:**
```hcl
module "prod_custom_cert" {
  source = "./path/to/digitaloceancertificate/v1/iac/tf"

  metadata = {
    name = "prod-custom-cert"
    id   = "docert-xyz789"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    certificate_name = "prod-custom-cert"
    type             = "custom"
    custom = {
      private_key       = file("${path.module}/secrets/privkey.pem")
      leaf_certificate  = file("${path.module}/secrets/cert.pem")
      certificate_chain = file("${path.module}/secrets/chain.pem")
    }
    description = "Custom EV certificate"
    tags        = ["env:production", "cert-type:ev"]
  }
}
```

### 3. Deploy

```bash
# Initialize Terraform (downloads providers)
terraform init

# Preview changes
terraform plan

# Apply changes
terraform apply
```

### 4. Retrieve Outputs

```bash
# Get certificate ID
terraform output certificate_id

# Get expiration timestamp
terraform output expiry_rfc3339

# Get all outputs as JSON
terraform output -json
```

## Module Variables

### `metadata` (Required)

Metadata for the certificate resource.

```hcl
metadata = {
  name    = string              # Human-readable name
  id      = optional(string)    # Unique ID (e.g., docert-abc123)
  org     = optional(string)    # Organization name
  env     = optional(string)    # Environment (dev, staging, prod)
  labels  = optional(map(string))  # Key-value labels
  tags    = optional(list(string)) # List of tags
  version = optional(object({   # Version metadata
    id      = string
    message = string
  }))
}
```

### `spec` (Required)

Certificate specification.

```hcl
spec = {
  certificate_name = string     # Unique certificate name (1-64 chars)
  type             = string     # "letsEncrypt" or "custom"

  # Let's Encrypt parameters (required when type = "letsEncrypt")
  lets_encrypt = optional(object({
    domains            = list(string)  # List of domains/wildcards
    disable_auto_renew = optional(bool, false)  # Disable auto-renewal
  }))

  # Custom certificate parameters (required when type = "custom")
  custom = optional(object({
    leaf_certificate  = string  # PEM-encoded public certificate
    private_key       = string  # PEM-encoded private key
    certificate_chain = optional(string)  # PEM-encoded intermediate chain
  }))

  # Optional fields
  description = optional(string)      # Max 128 characters
  tags        = optional(list(string))  # Unique tags
}
```

## Module Outputs

| Output | Type | Description |
|--------|------|-------------|
| `certificate_id` | string | DigitalOcean certificate UUID |
| `expiry_rfc3339` | string | Certificate expiration timestamp (RFC 3339) |
| `certificate_name` | string | Name of the certificate |
| `certificate_type` | string | Type of certificate (lets_encrypt or custom) |
| `certificate_state` | string | Certificate state (verified, pending, error) |

## Discriminated Union Implementation

The module implements the discriminated union pattern using Terraform's conditional logic:

```hcl
# In locals.tf
locals {
  cert_type = lower(var.spec.type) == "letsencrypt" ? "lets_encrypt" : "custom"
  is_lets_encrypt = lower(var.spec.type) == "letsencrypt"
  is_custom       = lower(var.spec.type) == "custom"
}

# In main.tf
resource "digitalocean_certificate" "certificate" {
  name = var.spec.certificate_name
  type = local.cert_type

  # Conditionally set fields based on certificate type
  domains           = local.is_lets_encrypt ? local.le_domains : null
  leaf_certificate  = local.is_custom ? local.custom_leaf_cert : null
  private_key       = local.is_custom ? local.custom_private_key : null
  certificate_chain = local.is_custom && local.custom_cert_chain != "" ? local.custom_cert_chain : null
}
```

This ensures type safety and prevents invalid configurations.

## Secret Management

### ⚠️ Security Warning

**Never commit private keys to Git!** Use one of the following secure methods:

### Option 1: Terraform Vault Provider

```hcl
data "vault_generic_secret" "cert" {
  path = "secret/digitalocean/prod-cert"
}

module "prod_cert" {
  source = "./path/to/module"
  
  spec = {
    type = "custom"
    custom = {
      private_key       = data.vault_generic_secret.cert.data["private_key"]
      leaf_certificate  = data.vault_generic_secret.cert.data["leaf_certificate"]
      certificate_chain = data.vault_generic_secret.cert.data["certificate_chain"]
    }
  }
}
```

### Option 2: AWS Secrets Manager

```hcl
data "aws_secretsmanager_secret_version" "cert" {
  secret_id = "digitalocean/prod-cert"
}

locals {
  cert_data = jsondecode(data.aws_secretsmanager_secret_version.cert.secret_string)
}

module "prod_cert" {
  source = "./path/to/module"
  
  spec = {
    type = "custom"
    custom = {
      private_key       = local.cert_data["private_key"]
      leaf_certificate  = local.cert_data["leaf_certificate"]
      certificate_chain = local.cert_data["certificate_chain"]
    }
  }
}
```

### Option 3: Terraform Variables (with encrypted backend)

```hcl
variable "cert_private_key" {
  type      = string
  sensitive = true
}

module "prod_cert" {
  source = "./path/to/module"
  
  spec = {
    type = "custom"
    custom = {
      private_key       = var.cert_private_key
      leaf_certificate  = var.cert_leaf
      certificate_chain = var.cert_chain
    }
  }
}
```

Set via environment variables:
```bash
export TF_VAR_cert_private_key="$(cat privkey.pem)"
export TF_VAR_cert_leaf="$(cat cert.pem)"
export TF_VAR_cert_chain="$(cat chain.pem)"
terraform apply
```

## Zero-Downtime Certificate Rotation

The module implements `create_before_destroy` lifecycle policy for custom certificates:

```hcl
resource "digitalocean_certificate" "certificate" {
  # ... configuration ...

  lifecycle {
    create_before_destroy = true
  }
}
```

**How it works:**
1. When certificate materials change, Terraform detects replacement is needed
2. **First**: Creates new certificate resource with new PEM content
3. **Second**: Updates resources referencing the certificate (e.g., Load Balancers)
4. **Third**: Deletes old certificate resource

This ensures HTTPS traffic is never interrupted during rotation.

## Integration with Load Balancers

Reference the certificate ID in DigitalOcean Load Balancer modules:

```hcl
module "prod_cert" {
  source = "./path/to/digitaloceancertificate/v1/iac/tf"
  # ... certificate config ...
}

resource "digitalocean_loadbalancer" "public" {
  name   = "prod-lb"
  region = "nyc3"

  forwarding_rule {
    entry_port       = 443
    entry_protocol   = "https"
    target_port      = 80
    target_protocol  = "http"
    certificate_id   = module.prod_cert.certificate_id
  }
}
```

## Examples

### Example 1: Minimal Let's Encrypt Certificate

```hcl
module "simple_cert" {
  source = "./path/to/module"

  metadata = {
    name = "simple-cert"
    env  = "dev"
  }

  spec = {
    certificate_name = "simple-cert"
    type             = "letsEncrypt"
    lets_encrypt = {
      domains = ["dev.example.com"]
    }
  }
}
```

### Example 2: Wildcard Certificate for Staging

```hcl
module "staging_wildcard" {
  source = "./path/to/module"

  metadata = {
    name = "staging-wildcard-cert"
    env  = "staging"
  }

  spec = {
    certificate_name = "staging-wildcard-cert"
    type             = "letsEncrypt"
    lets_encrypt = {
      domains = [
        "staging.example.com",
        "*.staging.example.com"
      ]
    }
    description = "Wildcard cert for all staging subdomains"
    tags        = ["env:staging", "cert-type:wildcard"]
  }
}
```

### Example 3: Custom Certificate with Secrets from Vault

```hcl
data "vault_generic_secret" "prod_cert" {
  path = "secret/digitalocean/prod-cert"
}

module "prod_custom_cert" {
  source = "./path/to/module"

  metadata = {
    name = "prod-custom-cert"
    env  = "production"
    org  = "acme-corp"
  }

  spec = {
    certificate_name = "prod-custom-cert"
    type             = "custom"
    custom = {
      private_key       = data.vault_generic_secret.prod_cert.data["private_key"]
      leaf_certificate  = data.vault_generic_secret.prod_cert.data["leaf_certificate"]
      certificate_chain = data.vault_generic_secret.prod_cert.data["certificate_chain"]
    }
    description = "Production EV certificate, expires 2026-01-15"
    tags        = ["env:production", "cert-type:ev"]
  }
}
```

## Troubleshooting

### Issue: "certificate name already exists"

**Cause**: Duplicate certificate name in DigitalOcean account.

**Solution**: Choose a unique name or delete the existing certificate:
```bash
doctl compute certificate list
doctl compute certificate delete <certificate-id>
```

### Issue: Let's Encrypt certificate stuck in "pending"

**Cause**: DNS is not managed by DigitalOcean.

**Solution**: Verify DNS zone exists in DigitalOcean:
```bash
doctl compute domain list
```

If DNS is external, use the Custom certificate workflow instead.

### Issue: "invalid PEM format"

**Cause**: Malformed PEM blocks.

**Solution**: Validate PEM format:
```bash
# Validate certificate
openssl x509 -in cert.pem -text -noout

# Validate private key
openssl rsa -in privkey.pem -check

# Verify key matches certificate
openssl x509 -noout -modulus -in cert.pem | openssl md5
openssl rsa -noout -modulus -in privkey.pem | openssl md5
# (The MD5 hashes should match)
```

### Issue: Browser shows "untrusted certificate"

**Cause**: Missing or incomplete certificate chain.

**Solution**: Always provide the full intermediate chain in `certificate_chain`:
```hcl
custom = {
  certificate_chain = file("${path.module}/fullchain.pem")  # Use fullchain.pem from CA
}
```

## Best Practices

### 1. Use Remote State

Store Terraform state in a secure, remote backend:

```hcl
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "digitalocean/certificates/prod.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
}
```

### 2. Enable State Locking

Prevent concurrent modifications:
- **S3 backend**: Use DynamoDB table for locking
- **Terraform Cloud**: Built-in locking
- **GitLab/GitHub**: CI/CD pipeline locking

### 3. Pin Provider Versions

```hcl
terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.34.0"  # Pin to specific version
    }
  }
}
```

### 4. Use Workspaces for Environments

```bash
terraform workspace new dev
terraform workspace new staging
terraform workspace new production

terraform workspace select production
terraform apply
```

### 5. Monitor Certificate Expiration

Set up DigitalOcean Uptime alerts or use external monitoring:

```bash
# Example: Prometheus Blackbox Exporter
- job_name: 'ssl_expiry'
  metrics_path: /probe
  params:
    module: [http_2xx]
  static_configs:
    - targets:
      - https://example.com
  relabel_configs:
    - source_labels: [__address__]
      target_label: __param_target
    - target_label: instance
      replacement: blackbox_exporter:9115
```

## Terraform Commands Reference

```bash
# Initialize module
terraform init

# Validate configuration
terraform validate

# Format code
terraform fmt -recursive

# Plan changes
terraform plan -out=tfplan

# Apply changes
terraform apply tfplan

# Show current state
terraform show

# List resources
terraform state list

# Show specific resource
terraform state show module.prod_cert.digitalocean_certificate.certificate

# Import existing certificate
terraform import module.prod_cert.digitalocean_certificate.certificate <certificate-id>

# Destroy resources
terraform destroy
```

## Further Reading

- **Component Overview**: See [../../README.md](../../README.md)
- **Comprehensive Guide**: See [../../docs/README.md](../../docs/README.md)
- **Examples**: See [../../examples.md](../../examples.md)
- **Terraform DigitalOcean Provider**: [Registry Docs](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/certificate)

## Support

For module-specific issues:
1. Enable debug logging: `export TF_LOG=DEBUG`
2. Check Terraform logs: `terraform apply 2>&1 | tee terraform.log`
3. Validate DigitalOcean API connectivity:
   ```bash
   curl -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" https://api.digitalocean.com/v2/account
   ```

For component issues, see the main [README.md](../../README.md).

