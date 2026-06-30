# DigitalOcean Droplet - Terraform Implementation

## Overview

This directory contains the Terraform implementation for provisioning DigitalOcean Droplets through the Planton API. The module reads a structured input matching the `DigitalOceanDropletSpec` protobuf specification and deploys the corresponding virtual machine using the DigitalOcean Terraform provider.

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
   - `spec`: Droplet specification (name, region, size, image, SSH keys, etc.)
   - `digitalocean_token`: DigitalOcean API authentication

2. **locals.tf**: Transforms input data:
   - Maps protobuf region enum to DigitalOcean region slugs
   - Extracts VPC UUID from spec
   - Filters and processes volume IDs
   - Combines user tags with metadata tags
   - Inverts disable_monitoring flag

3. **main.tf**: Provisions infrastructure:
   - Creates `digitalocean_droplet` resource
   - Applies SSH keys, VPC, volumes, tags, and user data
   - Configures optional features (IPv6, backups, monitoring)

4. **outputs.tf**: Exports stack outputs:
   - Droplet ID, IPv4/IPv6 addresses, VPC UUID
   - Status, URN, tags, and image ID

## Prerequisites

### Required Tools

- **Terraform 1.0+**: Install from https://www.terraform.io/downloads
- **DigitalOcean Account**: Sign up at https://cloud.digitalocean.com
- **DigitalOcean API Token**: Generate from https://cloud.digitalocean.com/account/api/tokens
- **SSH Key**: Upload to DigitalOcean or create via API

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
module "droplet" {
  source = "./iac/tf"

  metadata = {
    name = "my-dev-server"
  }

  spec = {
    droplet_name = "dev-server-01"
    region       = "digital_ocean_region_nyc3"
    size         = "s-2vcpu-4gb"
    image        = "ubuntu-22-04-x64"
    ssh_keys     = ["your-ssh-key-fingerprint"]
    vpc = {
      value = "vpc-uuid-for-dev"
    }
    tags = ["development", "web"]
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
output "droplet_ipv4" {
  value = module.droplet.ipv4_address
}

output "droplet_id" {
  value = module.droplet.droplet_id
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
- Droplet will be created with correct specifications
- SSH keys, VPC, and tags are configured
- No unexpected changes

#### 5. Apply Configuration

```bash
terraform apply
```

Type `yes` when prompted to confirm.

#### 6. Verify Outputs

```bash
terraform output droplet_ipv4
terraform output droplet_id
```

## Advanced Usage

### Production Droplet with Backups and Monitoring

```hcl
module "prod_droplet" {
  source = "./iac/tf"

  metadata = {
    name = "prod-web-01"
  }

  spec = {
    droplet_name = "prod-web-01"
    region       = "digital_ocean_region_nyc3"
    size         = "g-4vcpu-16gb"
    image        = "ubuntu-22-04-x64"
    ssh_keys = [
      "prod-deploy-key-fingerprint",
      "ops-team-key-fingerprint"
    ]
    vpc = {
      value = "vpc-uuid-prod"
    }
    enable_backups     = true
    disable_monitoring = false
    tags = [
      "production",
      "web-tier",
      "prod-firewall"
    ]
  }

  digitalocean_token = var.digitalocean_token
}
```

### Droplet with Block Storage Volumes

```hcl
module "droplet_with_volume" {
  source = "./iac/tf"

  metadata = {
    name = "app-server"
  }

  spec = {
    droplet_name = "app-server-01"
    region       = "digital_ocean_region_sfo3"
    size         = "g-2vcpu-8gb"
    image        = "ubuntu-22-04-x64"
    ssh_keys     = ["deploy-key-fingerprint"]
    vpc = {
      value = "vpc-uuid"
    }
    volume_ids = [
      {
        value = "volume-uuid-1"
      },
      {
        value = "volume-uuid-2"
      }
    ]
    enable_backups = true
    tags           = ["production", "app"]
  }

  digitalocean_token = var.digitalocean_token
}
```

### Droplet with Cloud-Init User Data

```hcl
module "automated_droplet" {
  source = "./iac/tf"

  metadata = {
    name = "automated-server"
  }

  spec = {
    droplet_name = "automated-server"
    region       = "digital_ocean_region_nyc3"
    size         = "s-2vcpu-4gb"
    image        = "ubuntu-22-04-x64"
    ssh_keys     = ["key-fingerprint"]
    vpc = {
      value = "vpc-uuid"
    }
    user_data = <<-EOT
      #cloud-config
      package_update: true
      packages:
        - nginx
        - fail2ban
      runcmd:
        - systemctl enable nginx
        - systemctl start nginx
    EOT
    tags = ["staging"]
  }

  digitalocean_token = var.digitalocean_token
}
```

### Droplet with IPv6 Enabled

```hcl
module "ipv6_droplet" {
  source = "./iac/tf"

  metadata = {
    name = "ipv6-server"
  }

  spec = {
    droplet_name = "ipv6-server"
    region       = "digital_ocean_region_nyc3"
    size         = "s-2vcpu-4gb"
    image        = "ubuntu-22-04-x64"
    ssh_keys     = ["key-fingerprint"]
    vpc = {
      value = "vpc-uuid"
    }
    enable_ipv6 = true
    tags        = ["ipv6-enabled"]
  }

  digitalocean_token = var.digitalocean_token
}
```

## Implementation Details

### Region Mapping

The module maps protobuf enum values to DigitalOcean region slugs:

| Protobuf Enum | Terraform Region Slug |
|---------------|----------------------|
| `digital_ocean_region_nyc1` | `nyc1` |
| `digital_ocean_region_nyc3` | `nyc3` |
| `digital_ocean_region_sfo3` | `sfo3` |
| `digital_ocean_region_ams3` | `ams3` |
| `digital_ocean_region_sgp1` | `sgp1` |
| `digital_ocean_region_lon1` | `lon1` |
| `digital_ocean_region_fra1` | `fra1` |

(Full mapping available in `locals.tf`)

### Tag Management

Tags are automatically combined from:
- User-provided tags in `spec.tags`
- Metadata tags (`managed-by`, `resource-kind`, `resource-name`)

Example:
```hcl
tags = [
  "production",                        # user-provided
  "managed-by:planton",       # auto-added
  "resource-kind:digitalocean-droplet",  # auto-added
  "resource-name:my-droplet"          # auto-added
]
```

### Monitoring Default

Monitoring is **enabled by default**. To disable:

```hcl
spec = {
  # ... other fields ...
  disable_monitoring = true
}
```

### Volume Attachment

Volumes must be created separately before referencing:

```hcl
resource "digitalocean_volume" "data" {
  name   = "app-data"
  region = "nyc3"
  size   = 100  # GB
}

module "droplet" {
  source = "./iac/tf"

  spec = {
    # ... other fields ...
    volume_ids = [
      {
        value = digitalocean_volume.data.id
      }
    ]
  }
}
```

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
    key                         = "digitaloceandroplet/my-droplet.tfstate"
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
| `droplet_id` | Droplet unique identifier | `"123456789"` |
| `ipv4_address` | Public IPv4 address | `"192.0.2.1"` |
| `ipv6_address` | Public IPv6 address (if enabled) | `"2001:db8::1"` |
| `ipv4_address_private` | Private IPv4 address (VPC) | `"10.116.0.2"` |
| `image_id` | Image slug or ID used | `"ubuntu-22-04-x64"` |
| `vpc_uuid` | VPC UUID | `"uuid"` |
| `urn` | DigitalOcean URN | `"do:droplet:123456789"` |
| `status` | Droplet status | `"active"` |
| `tags` | All applied tags | `["production", "managed-by:planton"]` |

Access outputs:

```bash
terraform output droplet_id
terraform output -json ipv4_address
```

## Troubleshooting

### Common Errors

#### Error: "Droplet name already exists"

**Cause**: Another Droplet with the same name exists in your account.

**Solution**: Either:
- Delete the existing Droplet
- Choose a different `droplet_name`
- Import the existing Droplet into Terraform state:
  ```bash
  terraform import module.droplet.digitalocean_droplet.droplet 123456789
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

#### Error: "SSH key not found"

**Cause**: SSH key fingerprint or ID doesn't exist in your DigitalOcean account.

**Solution**:
- Verify SSH keys in DigitalOcean control panel
- List SSH keys via CLI: `doctl compute ssh-key list`
- Use correct fingerprint format (MD5 or SHA256)

#### Error: "VPC not found"

**Cause**: VPC UUID is incorrect or VPC doesn't exist in the specified region.

**Solution**:
- Verify VPC UUID in DigitalOcean control panel
- Ensure VPC exists in the same region as Droplet
- List VPCs: `doctl vpcs list`

#### Error: "Volume not found or already attached"

**Cause**: Volume ID is incorrect, or volume is already attached to another Droplet.

**Solution**:
- Verify volume exists: `doctl compute volume list`
- Ensure volume is in the same region as Droplet
- Detach volume from other Droplet if needed

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
terraform state show 'module.droplet.digitalocean_droplet.droplet'

# View entire state
terraform show
```

### SSH Access Issues

**Cannot SSH to Droplet**:

1. Verify SSH key is correctly added to Droplet:
   ```bash
   ssh -i ~/.ssh/id_rsa root@<droplet-ip>
   ```

2. Check Cloud Firewall rules allow SSH (port 22)

3. Verify Droplet is in "active" status:
   ```bash
   terraform output status
   ```

4. Check SSH service is running (via DigitalOcean console):
   ```bash
   systemctl status sshd
   ```

## Advanced Topics

### Importing Existing Droplets

If you have existing DigitalOcean Droplets, import them into Terraform:

```bash
terraform import 'module.droplet.digitalocean_droplet.droplet' 123456789
```

Where `123456789` is the Droplet ID from DigitalOcean.

### Using Data Sources

Reference existing DigitalOcean resources:

```hcl
# Look up an existing VPC
data "digitalocean_vpc" "existing" {
  name = "production-vpc"
}

module "droplet" {
  source = "./iac/tf"

  spec = {
    # ... other fields ...
    vpc = {
      value = data.digitalocean_vpc.existing.id
    }
  }
}
```

### Module Composition

Combine with other Terraform modules:

```hcl
# Create a VPC
module "vpc" {
  source = "../digitaloceand-vpc"
  # ...
}

# Create a Volume
resource "digitalocean_volume" "data" {
  name   = "app-data"
  region = "nyc3"
  size   = 100
}

# Create Droplet using both
module "droplet" {
  source = "../digitaloceandroplet"

  spec = {
    # ... other fields ...
    vpc = {
      value = module.vpc.vpc_id
    }
    volume_ids = [
      {
        value = digitalocean_volume.data.id
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

### Droplet Provisioning Time

- **Standard Droplet**: ~55 seconds from API call to active
- **With Volumes**: +10 seconds for attachment
- **With user_data**: +variable (depends on script complexity)

### API Rate Limits

DigitalOcean API limits:
- **5,000 requests per hour** per account
- Terraform operations are well within limits for normal usage

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

## Security Best Practices

1. **Always use SSH keys**: Never rely on password authentication
2. **Enable monitoring**: Free and provides essential metrics
3. **Use VPCs**: Isolate Droplets from public internet
4. **Enable backups for production**: Disaster recovery insurance
5. **Restrict SSH access**: Use Cloud Firewalls to limit SSH to known IPs
6. **Keep OS updated**: Use cloud-init to enable automated security updates
7. **Rotate SSH keys**: Periodically update authorized keys
8. **Use fail2ban**: Automatically ban IPs with failed SSH attempts

## References

- **Terraform DigitalOcean Provider**: https://registry.terraform.io/providers/digitalocean/digitalocean/latest/docs
- **DigitalOcean Droplets API**: https://docs.digitalocean.com/reference/api/api-reference/#tag/Droplets
- **Terraform Documentation**: https://www.terraform.io/docs
- **Planton Docs**: ../../docs/README.md
- **Examples**: ../../examples.md

## Best Practices

1. **Version Pin Providers**: Always use version constraints in `provider.tf`
2. **Use Remote State**: For team collaboration, use remote backends
3. **Modularize**: Keep Droplet configuration separate from other infrastructure
4. **Validate First**: Run `terraform validate` and `terraform plan` before `apply`
5. **Never Commit State**: Add `terraform.tfstate` to `.gitignore`
6. **Document Changes**: Use comments in HCL files for complex configurations

## Contributing

When modifying this module:

1. Maintain backward compatibility for existing specs
2. Update variable descriptions in `variables.tf`
3. Add new outputs to `outputs.tf` if exposing new values
4. Test with both minimal and complex configurations
5. Update this README for significant changes

## License

This implementation is part of the Planton monorepo and follows the same license.

