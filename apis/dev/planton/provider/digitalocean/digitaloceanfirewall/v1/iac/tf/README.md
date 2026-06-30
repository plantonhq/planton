# DigitalOcean Firewall - Terraform Module

This Terraform module provisions and manages stateful, network-edge firewalls for DigitalOcean Droplets using the official `digitalocean` provider.

## Overview

The module implements Planton's `DigitalOceanFirewall` protobuf spec, providing a declarative interface for:

- **Inbound Rules**: Control traffic allowed *to* Droplets
- **Outbound Rules**: Control traffic allowed *from* Droplets
- **Tag-Based Targeting**: Production-standard, scales infinitely, auto-scaling friendly
- **Static Droplet IDs**: Dev/testing only, max 10 Droplets
- **Resource-Aware Rules**: Support for Load Balancer UIDs, Kubernetes cluster IDs

## Module Structure

```
iac/tf/
├── variables.tf    # Input variable definitions
├── locals.tf       # Local value transformations
├── main.tf         # Firewall resource definition
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
cd apis/dev/planton/provider/digitalocean/digitaloceanfirewall/v1/iac/tf

# Set your DigitalOcean API token
export DIGITALOCEAN_TOKEN="your-digitalocean-api-token"
```

### 2. Create Terraform Configuration

**Production Web Tier Firewall:**
```hcl
module "prod_web_firewall" {
  source = "./path/to/digitaloceanfirewall/v1/iac/tf"

  metadata = {
    name = "prod-web-firewall"
    id   = "dofw-abc123"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    name = "prod-web-firewall"
    tags = ["prod-web-tier"]

    inbound_rules = [
      {
        protocol                   = "tcp"
        port_range                 = "443"
        source_load_balancer_uids  = ["lb-abc123"]
      },
      {
        protocol         = "tcp"
        port_range       = "22"
        source_addresses = ["203.0.113.10/32"]
      }
    ]

    outbound_rules = [
      {
        protocol          = "tcp"
        port_range        = "5432"
        destination_tags  = ["prod-db-tier"]
      },
      {
        protocol                = "tcp"
        port_range              = "443"
        destination_addresses   = ["0.0.0.0/0", "::/0"]
      },
      {
        protocol                = "udp"
        port_range              = "53"
        destination_addresses   = ["0.0.0.0/0", "::/0"]
      }
    ]
  }
}
```

**Production Database Tier Firewall:**
```hcl
module "prod_db_firewall" {
  source = "./path/to/digitaloceanfirewall/v1/iac/tf"

  metadata = {
    name = "prod-db-firewall"
    env  = "production"
  }

  spec = {
    name = "prod-db-firewall"
    tags = ["prod-db-tier"]

    inbound_rules = [
      {
        protocol     = "tcp"
        port_range   = "5432"
        source_tags  = ["prod-web-tier"]
      },
      {
        protocol         = "tcp"
        port_range       = "22"
        source_addresses = ["203.0.113.10/32"]
      }
    ]

    outbound_rules = [
      {
        protocol                = "tcp"
        port_range              = "443"
        destination_addresses   = ["91.189.88.0/21"]  # Ubuntu repos
      },
      {
        protocol                = "udp"
        port_range              = "53"
        destination_addresses   = ["1.1.1.1/32"]  # Cloudflare DNS
      }
    ]
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
# Get firewall ID
terraform output firewall_id

# Get all outputs as JSON
terraform output -json
```

## Module Variables

### `metadata` (Required)

Metadata for the firewall resource.

```hcl
metadata = {
  name    = string              # Human-readable name
  id      = optional(string)    # Unique ID (e.g., dofw-abc123)
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

Firewall specification.

```hcl
spec = {
  name = string     # Firewall name (1-255 chars, unique per account)

  # Inbound rules (traffic allowed *to* Droplets)
  inbound_rules = optional(list(object({
    protocol                   = string              # "tcp", "udp", or "icmp"
    port_range                 = optional(string)    # "22", "8000-9000", "1-65535"
    source_addresses           = optional(list(string))   # CIDR blocks
    source_droplet_ids         = optional(list(number))   # Droplet IDs
    source_tags                = optional(list(string))   # Droplet tags
    source_kubernetes_ids      = optional(list(string))   # K8s cluster IDs
    source_load_balancer_uids  = optional(list(string))   # LB UIDs
  })), [])

  # Outbound rules (traffic allowed *from* Droplets)
  outbound_rules = optional(list(object({
    protocol                        = string              # "tcp", "udp", or "icmp"
    port_range                      = optional(string)    # "443", "5432", etc.
    destination_addresses           = optional(list(string))   # CIDR blocks
    destination_droplet_ids         = optional(list(number))   # Droplet IDs
    destination_tags                = optional(list(string))   # Droplet tags
    destination_kubernetes_ids      = optional(list(string))   # K8s cluster IDs
    destination_load_balancer_uids  = optional(list(string))   # LB UIDs
  })), [])

  # The Droplet IDs to which this firewall is applied (max 10, dev only)
  droplet_ids = optional(list(number), [])

  # The names of Droplet tags to which this firewall is applied (production)
  tags = optional(list(string), [])
}
```

## Module Outputs

| Output | Type | Description |
|--------|------|-------------|
| `firewall_id` | string | DigitalOcean firewall ID |
| `firewall_name` | string | Name of the firewall |
| `firewall_status` | string | Status of the firewall |
| `created_at` | string | Timestamp when firewall was created |
| `pending_changes` | list | List of pending changes to the firewall |

## Production Best Practices

### 1. Tag-Based Targeting (Production Standard)

**✅ Do (Production):**
```hcl
spec = {
  tags = ["prod-web-tier", "all-instances"]
  # No static droplet_ids
}
```

**Advantages:**
- Unlimited Droplets (vs. 10-Droplet limit for static IDs)
- Auto-scaling friendly (new Droplets automatically protected)
- Composable (multiple firewalls via multiple tags)

**❌ Don't (Production Anti-Pattern):**
```hcl
spec = {
  droplet_ids = [123456, 234567, 345678]  # Breaks at 11 Droplets, not auto-scalable
}
```

### 2. Never Expose Management Ports to Internet

**✅ Do:**
```hcl
inbound_rules = [
  {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["203.0.113.10/32"]  # Office bastion only
  }
]
```

**❌ Don't:**
```hcl
inbound_rules = [
  {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = ["0.0.0.0/0"]  # NEVER in production!
  }
]
```

### 3. Use Load Balancer UIDs for Public Services

**✅ Do:**
```hcl
inbound_rules = [
  {
    protocol                   = "tcp"
    port_range                 = "443"
    source_load_balancer_uids  = ["lb-abc123"]  # Force traffic through LB
  }
]
```

**❌ Don't:**
```hcl
inbound_rules = [
  {
    protocol         = "tcp"
    port_range       = "443"
    source_addresses = ["0.0.0.0/0"]  # Bypasses LB, no centralized logging/SSL
  }
]
```

### 4. Implement Explicit Outbound Rules for High-Security Tiers

**✅ Do (Database Tier):**
```hcl
outbound_rules = [
  {
    protocol                = "tcp"
    port_range              = "443"
    destination_addresses   = ["91.189.88.0/21"]  # Specific Ubuntu repos
  },
  {
    protocol                = "udp"
    port_range              = "53"
    destination_addresses   = ["1.1.1.1/32"]  # Specific DNS resolver
  }
]
```

**❌ Don't (Database Tier):**
```hcl
outbound_rules = [
  {
    protocol                = "tcp"
    port_range              = "1-65535"
    destination_addresses   = ["0.0.0.0/0"]  # Too permissive for DB tier
  }
]
```

## Multi-Tier Architecture Example

### Web Tier Firewall
```hcl
module "web_tier_firewall" {
  source = "./path/to/module"

  metadata = {
    name = "web-tier-firewall"
    env  = "production"
  }

  spec = {
    name = "web-tier-firewall"
    tags = ["web-tier"]

    inbound_rules = [
      {
        protocol                   = "tcp"
        port_range                 = "443"
        source_load_balancer_uids  = ["lb-prod"]
      },
      {
        protocol         = "tcp"
        port_range       = "22"
        source_addresses = ["203.0.113.10/32"]
      }
    ]

    outbound_rules = [
      {
        protocol          = "tcp"
        port_range        = "5432"
        destination_tags  = ["db-tier"]
      },
      {
        protocol          = "tcp"
        port_range        = "6379"
        destination_tags  = ["cache-tier"]
      },
      {
        protocol                = "tcp"
        port_range              = "443"
        destination_addresses   = ["0.0.0.0/0"]
      },
      {
        protocol                = "udp"
        port_range              = "53"
        destination_addresses   = ["0.0.0.0/0"]
      }
    ]
  }
}
```

### Cache Tier Firewall
```hcl
module "cache_tier_firewall" {
  source = "./path/to/module"

  metadata = {
    name = "cache-tier-firewall"
    env  = "production"
  }

  spec = {
    name = "cache-tier-firewall"
    tags = ["cache-tier"]

    inbound_rules = [
      {
        protocol     = "tcp"
        port_range   = "6379"
        source_tags  = ["web-tier"]
      },
      {
        protocol         = "tcp"
        port_range       = "22"
        source_addresses = ["203.0.113.10/32"]
      }
    ]

    outbound_rules = [
      {
        protocol                = "tcp"
        port_range              = "443"
        destination_addresses   = ["0.0.0.0/0"]
      },
      {
        protocol                = "udp"
        port_range              = "53"
        destination_addresses   = ["0.0.0.0/0"]
      }
    ]
  }
}
```

### Database Tier Firewall
```hcl
module "db_tier_firewall" {
  source = "./path/to/module"

  metadata = {
    name = "db-tier-firewall"
    env  = "production"
  }

  spec = {
    name = "db-tier-firewall"
    tags = ["db-tier"]

    inbound_rules = [
      {
        protocol     = "tcp"
        port_range   = "5432"
        source_tags  = ["web-tier"]
      },
      {
        protocol         = "tcp"
        port_range       = "22"
        source_addresses = ["203.0.113.10/32"]
      }
    ]

    outbound_rules = [
      {
        protocol                = "tcp"
        port_range              = "443"
        destination_addresses   = ["91.189.88.0/21"]
      },
      {
        protocol                = "udp"
        port_range              = "53"
        destination_addresses   = ["1.1.1.1/32", "1.0.0.1/32"]
      }
    ]
  }
}
```

## The "Double Firewall" Reality

DigitalOcean Cloud Firewalls operate at the **network edge**, outside Droplets. Host-based firewalls (`ufw`, `iptables`) run **inside** the operating system. They are **independent**.

**Traffic Flow:**
```
Internet → Cloud Firewall (network edge) → Host Firewall (OS) → Application
```

### Common Pitfall

You configure Cloud Firewall to allow port 443, but connection times out because `ufw` inside Droplet is still blocking it.

### Solutions

**Option 1: Configure Both Firewalls** (Defense in Depth)
```bash
# Cloud Firewall: Allow HTTPS (via Terraform, this module)
# Host Firewall: Allow HTTPS
sudo ufw allow 443/tcp
sudo ufw enable
```

**Option 2: Disable Host Firewall** (Simplified Operations)
```bash
sudo ufw disable
```

For production, **Option 1** (defense in depth) is preferred, but **Option 2** (simplified ops) is common.

## Troubleshooting

### Issue: Connection Times Out Despite Firewall Rule

**Cause**: Host firewall (`ufw` or `iptables`) inside Droplet is blocking traffic.

**Solution**: Check host firewall via Web Console:
```bash
sudo ufw status verbose
```

Either allow the port in `ufw` or disable host firewall:
```bash
sudo ufw allow 22/tcp
# OR
sudo ufw disable
```

### Issue: Auto-Scaling Creates Unprotected Droplets

**Cause**: Using static `droplet_ids` instead of tags.

**Solution**: Switch to tag-based targeting:
```hcl
spec = {
  tags = ["web-tier"]
  # Remove droplet_ids
}
```

Apply tag to auto-scaling template.

### Issue: Firewall Applied But No Effect

**Cause**: Tag mismatch. Firewall has `tags = ["web-tier"]`, Droplets have tag `web-server`.

**Solution**: Verify tags match exactly. Check Droplet's "Networking" tab in DigitalOcean dashboard.

### Issue: "too many droplet_ids"

**Cause**: Static Droplet IDs hit 10-Droplet maximum.

**Solution**: Switch to tag-based targeting (unlimited).

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
terraform state show module.prod_web_firewall.digitalocean_firewall.firewall

# Import existing firewall
terraform import module.prod_web_firewall.digitalocean_firewall.firewall <firewall-id>

# Destroy resources
terraform destroy
```

## Best Practices

### 1. Use Remote State

Store Terraform state in a secure, remote backend:

```hcl
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "digitalocean/firewalls/prod.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
}
```

### 2. Pin Provider Versions

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

### 3. Use Workspaces for Environments

```bash
terraform workspace new dev
terraform workspace new staging
terraform workspace new production

terraform workspace select production
terraform apply
```

### 4. Separate Firewalls by Role

Instead of one monolithic firewall, create focused firewalls:
- **management-firewall** (SSH, monitoring) → Applied to all instances
- **web-tier-firewall** (HTTPS from LB) → Applied to web tier only
- **db-tier-firewall** (PostgreSQL from web tier) → Applied to DB tier only

This provides separation of concerns and composability.

## Further Reading

- **Component Overview**: See [../../README.md](../../README.md)
- **Comprehensive Guide**: See [../../docs/README.md](../../docs/README.md)
- **Examples**: See [../../examples.md](../../examples.md)
- **Terraform DigitalOcean Provider**: [Registry Docs](https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs/resources/firewall)

## Support

For module-specific issues:
1. Enable debug logging: `export TF_LOG=DEBUG`
2. Check Terraform logs: `terraform apply 2>&1 | tee terraform.log`
3. Validate DigitalOcean API connectivity:
   ```bash
   curl -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" https://api.digitalocean.com/v2/account
   ```

For component issues, see the main [README.md](../../README.md).

