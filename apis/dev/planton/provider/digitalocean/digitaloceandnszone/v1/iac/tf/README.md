# DigitalOcean DNS Zone - Terraform Implementation

## Overview

This directory contains the Terraform implementation for provisioning DigitalOcean DNS Zones through the Planton API. The module reads a structured input matching the `DigitalOceanDnsZone` protobuf specification and deploys the corresponding DNS zone and records using the DigitalOcean Terraform provider.

## Architecture

### Module Structure

```
iac/tf/
├── variables.tf    # Input variable definitions
├── provider.tf     # Terraform and provider version constraints
├── locals.tf       # Local values and transformations
├── main.tf         # Resource definitions
├── outputs.tf      # Output values
└── README.md       # This file
```

### Component Flow

1. **variables.tf**: Defines input variables:
   - `metadata`: Resource metadata (name, labels, tags)
   - `spec`: DNS zone specification (domain name, records)
   - `digitalocean_token`: DigitalOcean API authentication

2. **locals.tf**: Transforms input data:
   - Maps protobuf enum values to Terraform record types
   - Flattens multi-value records for iteration
   - Generates unique keys for each DNS record instance

3. **main.tf**: Provisions infrastructure:
   - Creates `digitalocean_domain` resource
   - Creates `digitalocean_record` resources (one per value)
   - Conditionally sets type-specific fields

4. **outputs.tf**: Exports stack outputs:
   - Zone name, ID, URN
   - DigitalOcean nameservers
   - All created DNS records

## Prerequisites

### Required Tools

- **Terraform 1.0+**: Install from https://www.terraform.io/downloads
- **DigitalOcean Account**: Sign up at https://cloud.digitalocean.com
- **DigitalOcean API Token**: Generate from https://cloud.digitalocean.com/account/api/tokens

### Provider Version

This module uses the DigitalOcean Terraform provider version **~> 2.0**:

```hcl
required_providers {
  digitalocean = {
    source  = "digitalocean/digitalocean"
    version = "~> 2.0"
  }
}
```

## Usage

### Basic Usage

#### 1. Create a Terraform Configuration

Create a root module that calls this module:

```hcl
# main.tf
module "dns_zone" {
  source = "./iac/tf"

  metadata = {
    name = "my-website-dns"
  }

  spec = {
    domain_name = "example.com"
    records = [
      {
        name = "@"
        type = "dns_record_type_a"
        values = [
          { value = "192.0.2.1" }
        ]
        ttl_seconds = 3600
      },
      {
        name = "www"
        type = "dns_record_type_cname"
        values = [
          { value = "example.com." }
        ]
        ttl_seconds = 3600
      }
    ]
  }

  digitalocean_token = var.digitalocean_token
}

# variables.tf
variable "digitalocean_token" {
  description = "DigitalOcean API token"
  type        = string
  sensitive   = true
}

# outputs.tf
output "zone_name" {
  value = module.dns_zone.zone_name
}

output "name_servers" {
  value = module.dns_zone.name_servers
}
```

#### 2. Set Authentication

```bash
export TF_VAR_digitalocean_token="dop_v1_xxxxxxxxxxxxxxxxxxxx"
```

Alternatively, create a `terraform.tfvars` file:

```hcl
digitalocean_token = "dop_v1_xxxxxxxxxxxxxxxxxxxx"
```

**Security Note**: Never commit `terraform.tfvars` to version control if it contains secrets.

#### 3. Initialize Terraform

```bash
terraform init
```

#### 4. Plan Deployment

```bash
terraform plan
```

Review the plan to verify:
- Domain will be created
- DNS records will be created
- No unexpected changes

#### 5. Apply Configuration

```bash
terraform apply
```

Type `yes` when prompted to confirm.

#### 6. Verify Outputs

```bash
terraform output zone_name
terraform output name_servers
```

### Advanced Usage

#### Email Configuration (MX Records with Priority)

```hcl
spec = {
  domain_name = "mycompany.com"
  records = [
    # Google Workspace MX records
    {
      name     = "@"
      type     = "dns_record_type_mx"
      priority = 1
      values   = [{ value = "aspmx.l.google.com." }]
    },
    {
      name     = "@"
      type     = "dns_record_type_mx"
      priority = 5
      values   = [{ value = "alt1.aspmx.l.google.com." }]
    },
    # SPF record
    {
      name = "@"
      type = "dns_record_type_txt"
      values = [
        { value = "v=spf1 include:_spf.google.com ~all" }
      ]
    }
  ]
}
```

#### CAA Records for Let's Encrypt

```hcl
spec = {
  domain_name = "secure-app.com"
  records = [
    {
      name  = "@"
      type  = "dns_record_type_caa"
      flags = 0
      tag   = "issue"
      values = [
        { value = "letsencrypt.org" }
      ]
    },
    {
      name  = "@"
      type  = "dns_record_type_caa"
      flags = 0
      tag   = "issuewild"
      values = [
        { value = "letsencrypt.org" }
      ]
    }
  ]
}
```

#### SRV Records for Service Discovery

```hcl
spec = {
  domain_name = "gaming-server.com"
  records = [
    {
      name     = "_minecraft._tcp"
      type     = "dns_record_type_srv"
      priority = 10
      weight   = 60
      port     = 25565
      values   = [{ value = "mc1.gaming-server.com." }]
    },
    {
      name = "mc1"
      type = "dns_record_type_a"
      values = [{ value = "192.0.2.10" }]
    }
  ]
}
```

### Multi-Environment Setup

Use Terraform workspaces for managing multiple environments:

```bash
# Development
terraform workspace new dev
terraform apply -var-file=dev.tfvars

# Production
terraform workspace new prod
terraform apply -var-file=prod.tfvars
```

**dev.tfvars**:
```hcl
digitalocean_token = "dop_v1_dev_token"
```

**prod.tfvars**:
```hcl
digitalocean_token = "dop_v1_prod_token"
```

## Implementation Details

### Record Type Mapping

The module maps protobuf enum values to DigitalOcean record types:

| Protobuf Enum | Terraform Type |
|---------------|----------------|
| `dns_record_type_a` | `A` |
| `dns_record_type_aaaa` | `AAAA` |
| `dns_record_type_cname` | `CNAME` |
| `dns_record_type_mx` | `MX` |
| `dns_record_type_txt` | `TXT` |
| `dns_record_type_srv` | `SRV` |
| `dns_record_type_caa` | `CAA` |
| `dns_record_type_ns` | `NS` |

This mapping is defined in `locals.tf`:

```hcl
locals {
  record_type_map = {
    "dns_record_type_a"    = "A"
    "dns_record_type_aaaa" = "AAAA"
    "dns_record_type_cname" = "CNAME"
    "dns_record_type_mx"   = "MX"
    "dns_record_type_txt"  = "TXT"
    "dns_record_type_srv"  = "SRV"
    "dns_record_type_caa"  = "CAA"
    "dns_record_type_ns"   = "NS"
  }
}
```

### Multi-Value Record Handling

When a record has multiple values (e.g., multiple MX servers), the module creates one `digitalocean_record` resource per value:

```hcl
locals {
  dns_records = flatten([
    for idx, record in coalesce(var.spec.records, []) : [
      for val_idx, value in record.values : {
        key = "${record.name}-${idx}-${val_idx}"
        # ... other fields
      }
    ]
  ])
}

resource "digitalocean_record" "dns_records" {
  for_each = { for record in local.dns_records : record.key => record }
  # ...
}
```

This allows each value to have independent priority (for MX records) and simplifies state management.

### Conditional Field Setting

The module conditionally sets fields based on record type:

```hcl
resource "digitalocean_record" "dns_records" {
  # ... basic fields ...

  priority = (
    each.value.type == "MX" || each.value.type == "SRV"
    ? coalesce(each.value.priority, 0)
    : null
  )

  weight = (
    each.value.type == "SRV"
    ? coalesce(each.value.weight, 0)
    : null
  )

  port = (
    each.value.type == "SRV"
    ? coalesce(each.value.port, 0)
    : null
  )

  flags = (
    each.value.type == "CAA"
    ? coalesce(each.value.flags, 0)
    : null
  )

  tag = (
    each.value.type == "CAA"
    ? each.value.tag
    : null
  )
}
```

This ensures only relevant fields are set, preventing API errors.

### Default Values

The module applies defaults:

- **TTL**: 3600 seconds (1 hour) if not specified
- **Priority**: 0 for MX/SRV if not specified
- **Weight**: 0 for SRV if not specified
- **Port**: 0 for SRV if not specified
- **Flags**: 0 for CAA if not specified

## State Management

### Local State (Development)

For local development, Terraform stores state in `terraform.tfstate`:

```bash
terraform apply
# Creates terraform.tfstate in current directory
```

**Warning**: Do not commit `terraform.tfstate` to version control (add to `.gitignore`).

### Remote State (Production)

For production, use a remote backend like DigitalOcean Spaces (S3-compatible):

```hcl
# backend.tf
terraform {
  backend "s3" {
    endpoint                    = "nyc3.digitaloceanspaces.com"
    region                      = "us-east-1"  # Dummy value required by S3 backend
    bucket                      = "my-terraform-state"
    key                         = "digitaloceandnszone/example.com.tfstate"
    skip_credentials_validation = true
    skip_metadata_api_check     = true
  }
}
```

Set credentials:

```bash
export AWS_ACCESS_KEY_ID="your-spaces-access-key"
export AWS_SECRET_ACCESS_KEY="your-spaces-secret-key"
```

Initialize backend:

```bash
terraform init -backend-config=backend.tf
```

## Outputs

The module exports the following outputs:

| Output | Description | Example |
|--------|-------------|---------|
| `zone_name` | Domain name | `"example.com"` |
| `zone_id` | DigitalOcean zone ID | `"example.com"` |
| `urn` | DigitalOcean URN | `"do:domain:example.com"` |
| `name_servers` | Nameservers for delegation | `["ns1.digitalocean.com", "ns2.digitalocean.com", "ns3.digitalocean.com"]` |
| `dns_records` | Map of all DNS records | `{ "@-0-0": { id = "123", fqdn = "example.com", ... } }` |

Access outputs:

```bash
terraform output zone_name
terraform output -json name_servers
terraform output -json dns_records
```

## Troubleshooting

### Common Errors

#### Error: "domain already exists"

**Cause**: The domain is already registered in your DigitalOcean account.

**Solution**: Import the existing domain into Terraform state:

```bash
terraform import module.dns_zone.digitalocean_domain.dns_zone example.com
```

#### Error: "Invalid authentication token"

**Cause**: The `digitalocean_token` variable is not set or is invalid.

**Solution**:

```bash
# Verify token is set
echo $TF_VAR_digitalocean_token

# Set it if empty
export TF_VAR_digitalocean_token="dop_v1_xxxxxxxxxxxxxxxxxxxx"

# Or use terraform.tfvars
echo 'digitalocean_token = "dop_v1_xxxxxxxxxxxxxxxxxxxx"' > terraform.tfvars
```

#### Error: "priority is required for MX records"

**Cause**: MX record missing priority field.

**Solution**: Add priority to your spec:

```hcl
{
  name     = "@"
  type     = "dns_record_type_mx"
  priority = 10  # Add this
  values   = [{ value = "mail.example.com." }]
}
```

#### Error: "record already exists"

**Cause**: Attempting to create a duplicate DNS record.

**Solution**: Check for duplicate records in your spec. Each unique combination of (name, type, value) should appear only once.

### Debugging

#### Enable Debug Logging

```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=terraform-debug.log
terraform apply
```

#### View Planned Changes

```bash
terraform plan -out=plan.tfplan
terraform show plan.tfplan
```

#### Inspect State

```bash
# List all resources
terraform state list

# Show specific resource
terraform state show 'digitalocean_record.dns_records["@-0-0"]'

# View entire state
terraform show
```

### DNS Propagation

After applying Terraform changes, DNS may take time to propagate:

**Immediate verification** (query DigitalOcean nameservers directly):
```bash
dig @ns1.digitalocean.com example.com
```

**Global propagation check**:
```bash
dig example.com @8.8.8.8
dig example.com @1.1.1.1
```

**Wait times**:
- DigitalOcean API: Immediate (seconds)
- Local ISP: TTL value (default 3600s = 1 hour)
- Global: 24-48 hours for full propagation

## Advanced Topics

### Importing Existing Resources

If you have existing DigitalOcean DNS zones, import them into Terraform:

```bash
# Import domain
terraform import 'digitalocean_domain.dns_zone' example.com

# Import DNS records (repeat for each record)
terraform import 'digitalocean_record.dns_records["www-0-0"]' example.com,123456789
```

Where `123456789` is the record ID from DigitalOcean API.

### Using Data Sources

Reference existing DigitalOcean resources:

```hcl
# Look up an existing Droplet
data "digitalocean_droplet" "web" {
  name = "web-server-01"
}

# Use its IP in DNS record
spec = {
  domain_name = "example.com"
  records = [
    {
      name   = "@"
      type   = "dns_record_type_a"
      values = [{ value = data.digitalocean_droplet.web.ipv4_address }]
    }
  ]
}
```

### Module Composition

Combine with other Terraform modules:

```hcl
# Create a Load Balancer
module "load_balancer" {
  source = "../digitalocean-load-balancer"
  # ...
}

# Point DNS to Load Balancer
module "dns_zone" {
  source = "../digitaloceandnszone"
  
  spec = {
    domain_name = "myapp.com"
    records = [
      {
        name   = "@"
        type   = "dns_record_type_a"
        values = [{ value = module.load_balancer.ip_address }]
      }
    ]
  }
}
```

### Terraform Validation

Validate configuration without applying:

```bash
terraform validate
```

Check formatting:

```bash
terraform fmt -check
terraform fmt -recursive  # Auto-format all files
```

## Performance Considerations

### API Rate Limits

DigitalOcean API limit: **250 requests/minute**

Terraform's default parallelism: **10 concurrent operations**

For large zones (100+ records), limit parallelism:

```bash
terraform apply -parallelism=5
```

### Large Zone Optimization

For zones with many records, use:

```bash
# Increase refresh interval to reduce API calls
terraform apply -refresh=false  # Skip refresh on apply

# Target specific resources
terraform apply -target='digitalocean_record.dns_records["www-0-0"]'
```

## Testing

### Plan-Only Mode

Test changes without applying:

```bash
terraform plan -out=plan.tfplan
# Review plan
# Do not apply
```

### Destroy and Recreate

Test full lifecycle:

```bash
terraform apply
terraform destroy
terraform apply  # Recreate from scratch
```

## References

- **Terraform DigitalOcean Provider**: https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs
- **DigitalOcean DNS API**: https://docs.digitalocean.com/reference/api/api-reference/#tag/Domains
- **Terraform Documentation**: https://www.terraform.io/docs
- **Planton Docs**: ../../docs/README.md
- **Examples**: ../../examples.md

## Best Practices

1. **Version Pin Providers**: Always use version constraints in `provider.tf`
2. **Use Remote State**: For team collaboration, use remote backends
3. **Modularize**: Keep DNS configuration separate from other infrastructure
4. **Validate First**: Run `terraform validate` and `terraform plan` before `apply`
5. **Commit State**: Never commit `terraform.tfstate` to version control
6. **Document Changes**: Use comments in HCL files for complex configurations

## Contributing

When modifying this module:

1. Maintain backward compatibility for existing specs
2. Update variable descriptions in `variables.tf`
3. Add new outputs to `outputs.tf` if exposing new values
4. Test with both simple (1 record) and complex (50+ records) zones
5. Update this README for significant changes

## License

This implementation is part of the Planton monorepo and follows the same license.

